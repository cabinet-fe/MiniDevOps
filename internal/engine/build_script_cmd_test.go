package engine

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsLegacyWindowsPowerShell(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, true},
		{`powershell`, true},
		{`C:\Program Files\PowerShell\7\pwsh.exe`, false},
		{`pwsh`, false},
	}
	for _, tc := range cases {
		if got := isLegacyWindowsPowerShell(tc.path); got != tc.want {
			t.Fatalf("isLegacyWindowsPowerShell(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestNewBuildScriptCommandBashUsesShOnUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	dir := t.TempDir()
	cmd, cleanup, err := newBuildScriptCommand(context.Background(), dir, "bash", "true")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	base := filepath.Base(cmd.Path)
	if base != "sh" {
		t.Fatalf("want sh, got path %q", cmd.Path)
	}
	if len(cmd.Args) < 3 || cmd.Args[1] != "-c" {
		t.Fatalf("unexpected args: %#v", cmd.Args)
	}
}

func TestNewBuildScriptCommandNode(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip()
	}
	dir := t.TempDir()
	cmd, cleanup, err := newBuildScriptCommand(context.Background(), dir, "node", "0")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if len(cmd.Args) < 3 || cmd.Args[1] != "-e" || cmd.Args[2] != "0" {
		t.Fatalf("unexpected args: %#v", cmd.Args)
	}
}

func TestFindPowerShellByTypeUnknown(t *testing.T) {
	_, err := findPowerShellByType("ps")
	if err == nil {
		t.Fatal("expected error for unknown script type")
	}
	if !strings.Contains(err.Error(), "未知 PowerShell 脚本类型") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFindPowerShellByTypeNotFound(t *testing.T) {
	t.Run("powershell", func(t *testing.T) {
		if _, err := exec.LookPath("powershell"); err == nil {
			t.Skip("powershell is installed")
		}
		_, err := findPowerShellByType("powershell")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "未找到 powershell") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("pwsh", func(t *testing.T) {
		if _, err := exec.LookPath("pwsh"); err == nil {
			t.Skip("pwsh is installed")
		}
		_, err := findPowerShellByType("pwsh")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "未找到 pwsh") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestNewBuildScriptCommandPwshUsesPwshBinary(t *testing.T) {
	pwshPath, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh not installed")
	}
	dir := t.TempDir()
	cmd, cleanup, err := newBuildScriptCommand(context.Background(), dir, "pwsh", "Write-Output ok")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if cmd.Path != pwshPath {
		t.Fatalf("want pwsh at %q, got %q", pwshPath, cmd.Path)
	}
}

func TestNewBuildScriptCommandPowerShellUsesPowerShellBinary(t *testing.T) {
	psPath, err := exec.LookPath("powershell")
	if err != nil {
		t.Skip("powershell not installed")
	}
	dir := t.TempDir()
	cmd, cleanup, err := newBuildScriptCommand(context.Background(), dir, "powershell", "Write-Output ok")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if cmd.Path != psPath {
		t.Fatalf("want powershell at %q, got %q", psPath, cmd.Path)
	}
}

func TestNewBuildScriptCommandPowerShellRejectsAndOnLegacy(t *testing.T) {
	psPath, err := exec.LookPath("powershell")
	if err != nil {
		t.Skip("powershell not installed")
	}
	if !isLegacyWindowsPowerShell(psPath) {
		t.Skip("powershell is not legacy Windows PowerShell 5.x")
	}
	_, _, err = newBuildScriptCommand(context.Background(), t.TempDir(), "powershell", "echo hi && echo bye")
	if err == nil {
		t.Fatal("expected error for && on legacy Windows PowerShell")
	}
	if !strings.Contains(err.Error(), "&&") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewBuildScriptCommandPwshAllowsAnd(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not installed")
	}
	dir := t.TempDir()
	_, cleanup, err := newBuildScriptCommand(context.Background(), dir, "pwsh", "Write-Output ok && Write-Output bye")
	if err != nil {
		t.Fatalf("pwsh should allow &&: %v", err)
	}
	defer cleanup()
}
