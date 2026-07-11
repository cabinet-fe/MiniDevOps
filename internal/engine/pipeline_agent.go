package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"buildflow/internal/model"
	"buildflow/internal/service"
)

// runMountedAgents 在构建产物成功后、分发前按序执行环境挂载的智能体。
// 单个失败仅写 ERROR 日志并继续，不将 Build.Status 改回 failed。
func (p *Pipeline) runMountedAgents(ctx context.Context, build *model.Build, project *model.Project, env *model.Environment, workDir string, writeLine func(string)) {
	if p.agentRepo == nil {
		return
	}
	agents, err := p.agentRepo.ListMountedAgentsForBuild(env.ID, project.ID)
	if err != nil {
		writeLine("ERROR: 加载挂载智能体失败: " + err.Error())
		return
	}
	if len(agents) == 0 {
		return
	}

	p.updateStageKeepSuccess(build, "agent")
	writeLine("=== Stage: Agents ===")

	envVars := os.Environ()
	if resolved, err := p.resolveEnvironmentVars(env.ID); err == nil {
		envVars = append(envVars, resolved...)
	} else {
		writeLine("WARNING: 解析环境变量失败，智能体将使用进程默认环境: " + err.Error())
	}

	proxySvc := service.NewAgentProxyService()
	for _, agent := range agents {
		if ctx.Err() != nil {
			writeLine("WARNING: 智能体阶段被取消")
			return
		}
		writeLine(fmt.Sprintf("--- Agent: %s (proxy=%s) ---", agent.Name, agent.ProxyKey))
		name, args, err := proxySvc.BuildRunCommand(agent.ProxyKey, agent.Prompt)
		if err != nil {
			writeLine("ERROR: " + err.Error())
			continue
		}
		if err := p.runAgentCommand(ctx, workDir, envVars, name, args, writeLine); err != nil {
			writeLine(fmt.Sprintf("ERROR: 智能体 %s 执行失败: %s", agent.Name, err.Error()))
			continue
		}
		writeLine(fmt.Sprintf("--- Agent %s finished ---", agent.Name))
	}
}

func (p *Pipeline) runAgentCommand(ctx context.Context, workDir string, envVars []string, name string, args []string, writeLine func(string)) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir
	cmd.Env = envVars
	configureBuildCmdProc(cmd)
	cmd.Cancel = func() error {
		return killBuildCmdProcess(cmd)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() { _ = killBuildCmdProcess(cmd) }()

	var scanWg sync.WaitGroup
	scanWg.Add(2)
	go func() {
		defer scanWg.Done()
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			writeLine(sc.Text())
		}
	}()
	go func() {
		defer scanWg.Done()
		sc := bufio.NewScanner(stderr)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			writeLine(sc.Text())
		}
	}()
	scanWg.Wait()
	return cmd.Wait()
}
