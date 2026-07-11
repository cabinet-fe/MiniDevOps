package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// AgentProxyInfo 本机 CLI 代理状态
type AgentProxyInfo struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Binary    string `json:"binary"`
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	Message   string `json:"message"`
}

type agentProxyDef struct {
	Key     string
	Name    string
	Binary  string
	Install func() (string, []string, error) // command + args; or shell via bash -c
	Upgrade func() (string, []string, error)
	// RunCmd 构建非交互执行命令：返回 name + args（不含 prompt，由调用方追加）
	RunCmd func(prompt string) (string, []string, error)
}

type AgentProxyService struct {
	lookPath func(string) (string, error)
	runCmd   func(ctx context.Context, name string, args ...string) (string, error)
}

func NewAgentProxyService() *AgentProxyService {
	return &AgentProxyService{
		lookPath: exec.LookPath,
		runCmd:   defaultRunCmd,
	}
}

func defaultRunCmd(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if out == "" {
		out = strings.TrimSpace(stderr.String())
	}
	return out, err
}

func agentProxyCatalog() []agentProxyDef {
	return []agentProxyDef{
		{
			Key:    "opencode",
			Name:   "OpenCode",
			Binary: "opencode",
			Install: func() (string, []string, error) {
				if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
					return "", nil, errors.New("仅支持 darwin/linux 安装")
				}
				return "bash", []string{"-c", "curl -fsSL https://opencode.ai/install | bash"}, nil
			},
			Upgrade: func() (string, []string, error) {
				return "opencode", []string{"upgrade"}, nil
			},
			RunCmd: func(prompt string) (string, []string, error) {
				return "opencode", []string{"run", prompt}, nil
			},
		},
		{
			Key:    "claude",
			Name:   "Claude Code",
			Binary: "claude",
			Install: func() (string, []string, error) {
				if _, err := exec.LookPath("npm"); err != nil {
					return "", nil, errors.New("未找到 npm，无法安装 Claude Code")
				}
				return "npm", []string{"i", "-g", "@anthropic-ai/claude-code"}, nil
			},
			Upgrade: func() (string, []string, error) {
				if _, err := exec.LookPath("npm"); err != nil {
					return "", nil, errors.New("未找到 npm，无法更新 Claude Code")
				}
				return "npm", []string{"i", "-g", "@anthropic-ai/claude-code@latest"}, nil
			},
			RunCmd: func(prompt string) (string, []string, error) {
				return "claude", []string{"-p", prompt, "--dangerously-skip-permissions"}, nil
			},
		},
		{
			Key:    "reasonix",
			Name:   "Reasonix",
			Binary: "reasonix",
			Install: func() (string, []string, error) {
				if _, err := exec.LookPath("npm"); err != nil {
					return "", nil, errors.New("未找到 npm，无法安装 reasonix")
				}
				return "npm", []string{"i", "-g", "reasonix"}, nil
			},
			Upgrade: func() (string, []string, error) {
				if _, err := exec.LookPath("npm"); err != nil {
					return "", nil, errors.New("未找到 npm，无法更新 reasonix")
				}
				return "npm", []string{"i", "-g", "reasonix@latest"}, nil
			},
			// reasonix 官方非交互子命令尚未稳定文档化；缺省返回明确错误
			RunCmd: func(_ string) (string, []string, error) {
				return "", nil, errors.New("reasonix 暂无已确认的非交互执行子命令，请查阅官方 CLI 文档后扩展")
			},
		},
	}
}

func findProxyDef(key string) (*agentProxyDef, error) {
	for _, d := range agentProxyCatalog() {
		if d.Key == key {
			def := d
			return &def, nil
		}
	}
	return nil, fmt.Errorf("未知代理: %s", key)
}

func (s *AgentProxyService) List() []AgentProxyInfo {
	defs := agentProxyCatalog()
	out := make([]AgentProxyInfo, 0, len(defs))
	for _, d := range defs {
		out = append(out, s.detect(d))
	}
	return out
}

func (s *AgentProxyService) detect(d agentProxyDef) AgentProxyInfo {
	info := AgentProxyInfo{
		Key:    d.Key,
		Name:   d.Name,
		Binary: d.Binary,
	}
	path, err := s.lookPath(d.Binary)
	if err != nil {
		info.Installed = false
		info.Message = "未安装"
		return info
	}
	info.Installed = true
	info.Path = path
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	version, err := s.runCmd(ctx, d.Binary, "--version")
	if err != nil {
		info.Message = "已安装，但获取版本失败: " + err.Error()
		return info
	}
	info.Version = firstLine(version)
	info.Message = "已安装"
	return info
}

func (s *AgentProxyService) Install(key string) (AgentProxyInfo, string, error) {
	def, err := findProxyDef(key)
	if err != nil {
		return AgentProxyInfo{}, "", err
	}
	name, args, err := def.Install()
	if err != nil {
		return AgentProxyInfo{}, "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	output, err := s.runCmd(ctx, name, args...)
	info := s.detect(*def)
	if err != nil {
		msg := output
		if msg == "" {
			msg = err.Error()
		}
		return info, msg, fmt.Errorf("安装失败: %s", msg)
	}
	// 刷新 PATH 后重新探测
	info = s.detect(*def)
	return info, output, nil
}

func (s *AgentProxyService) Upgrade(key string) (AgentProxyInfo, string, error) {
	def, err := findProxyDef(key)
	if err != nil {
		return AgentProxyInfo{}, "", err
	}
	if _, err := s.lookPath(def.Binary); err != nil {
		return AgentProxyInfo{}, "", errors.New("尚未安装，请先安装")
	}
	name, args, err := def.Upgrade()
	if err != nil {
		return AgentProxyInfo{}, "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	output, err := s.runCmd(ctx, name, args...)
	info := s.detect(*def)
	if err != nil {
		msg := output
		if msg == "" {
			msg = err.Error()
		}
		return info, msg, fmt.Errorf("更新失败: %s", msg)
	}
	info = s.detect(*def)
	return info, output, nil
}

// BuildRunCommand 返回在 workDir 下执行智能体的命令名与参数
func (s *AgentProxyService) BuildRunCommand(proxyKey, prompt string) (string, []string, error) {
	def, err := findProxyDef(proxyKey)
	if err != nil {
		return "", nil, err
	}
	if def.RunCmd == nil {
		return "", nil, errors.New("该代理未配置执行命令")
	}
	return def.RunCmd(prompt)
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
