package engine

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
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
