package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// newBuildScriptCommand constructs the build script command for the given type.
// cleanup removes any temporary script file; call it after cmd.Wait returns.
func newBuildScriptCommand(ctx context.Context, workDir, scriptType, script string) (cmd *exec.Cmd, cleanup func(), err error) {
	cleanup = func() {}
	st := strings.ToLower(strings.TrimSpace(scriptType))
	if st == "" {
		st = "bash"
	}

	switch st {
	case "node":
		return exec.CommandContext(ctx, "node", "-e", script), cleanup, nil
	case "python":
		return exec.CommandContext(ctx, "python3", "-c", script), cleanup, nil
	case "powershell", "pwsh":
		return newPowerShellBuildCommand(ctx, workDir, script)
	case "cmd", "batch":
		return newCmdBuildCommand(ctx, workDir, script)
	default:
		return newPOSIXShellBuildCommand(ctx, script)
	}
}

func newPOSIXShellBuildCommand(ctx context.Context, script string) (*exec.Cmd, func(), error) {
	if runtime.GOOS == "windows" {
		for _, name := range []string{"bash", "sh"} {
			path, err := exec.LookPath(name)
			if err == nil {
				return exec.CommandContext(ctx, path, "-c", script), func() {}, nil
			}
		}
		return nil, func() {}, fmt.Errorf("未找到 bash 或 sh。请在 Windows 上安装 Git for Windows，或将脚本类型改为 PowerShell / CMD")
	}
	return exec.CommandContext(ctx, "sh", "-c", script), func() {}, nil
}

func findPowerShell() (string, error) {
	for _, name := range []string{"powershell", "pwsh"} {
		path, err := exec.LookPath(name)
		if err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("未找到 powershell 或 pwsh")
}

func newPowerShellBuildCommand(ctx context.Context, workDir, script string) (*exec.Cmd, func(), error) {
	ps, err := findPowerShell()
	if err != nil {
		return nil, func() {}, err
	}
	f, err := os.CreateTemp(workDir, ".buildflow-*.ps1")
	if err != nil {
		return nil, func() {}, err
	}
	path := f.Name()
	if _, err := f.WriteString(script); err != nil {
		f.Close()
		os.Remove(path)
		return nil, func() {}, err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return nil, func() {}, err
	}
	cleanup := func() { os.Remove(path) }
	cmd := exec.CommandContext(ctx, ps,
		"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", path)
	return cmd, cleanup, nil
}

func newCmdBuildCommand(ctx context.Context, workDir, script string) (*exec.Cmd, func(), error) {
	cmdexe := os.Getenv("ComSpec")
	if cmdexe == "" {
		if windir := os.Getenv("SystemRoot"); windir != "" {
			cmdexe = filepath.Join(windir, "System32", "cmd.exe")
		}
	}
	if cmdexe == "" {
		if p, err := exec.LookPath("cmd"); err == nil {
			cmdexe = p
		} else {
			return nil, func() {}, fmt.Errorf("未找到 cmd.exe")
		}
	} else if _, err := os.Stat(cmdexe); err != nil {
		if p, err2 := exec.LookPath("cmd"); err2 == nil {
			cmdexe = p
		} else {
			return nil, func() {}, fmt.Errorf("未找到 cmd.exe")
		}
	}

	f, err := os.CreateTemp(workDir, ".buildflow-*.cmd")
	if err != nil {
		return nil, func() {}, err
	}
	path := f.Name()
	if _, err := f.WriteString(script); err != nil {
		f.Close()
		os.Remove(path)
		return nil, func() {}, err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return nil, func() {}, err
	}
	cleanup := func() { os.Remove(path) }
	cmd := exec.CommandContext(ctx, cmdexe, "/C", path)
	return cmd, cleanup, nil
}
