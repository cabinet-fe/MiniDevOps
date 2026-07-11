package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestAgentProxyDetectInstalled(t *testing.T) {
	s := NewAgentProxyService()
	s.lookPath = func(file string) (string, error) {
		if file == "opencode" {
			return "/usr/local/bin/opencode", nil
		}
		return "", errors.New("not found")
	}
	s.runCmd = func(ctx context.Context, name string, args ...string) (string, error) {
		if name == "opencode" && len(args) > 0 && args[0] == "--version" {
			return "opencode 1.2.3\nextra", nil
		}
		return "", errors.New("unexpected")
	}

	list := s.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 proxies, got %d", len(list))
	}
	var oc *AgentProxyInfo
	for i := range list {
		if list[i].Key == "opencode" {
			oc = &list[i]
		}
	}
	if oc == nil || !oc.Installed || oc.Version != "opencode 1.2.3" {
		t.Fatalf("unexpected opencode info: %+v", oc)
	}
	if filepath.Base(oc.Path) != "opencode" {
		t.Fatalf("unexpected path: %s", oc.Path)
	}
}

func TestAgentProxyDetectMissing(t *testing.T) {
	s := NewAgentProxyService()
	s.lookPath = func(string) (string, error) { return "", errors.New("not found") }
	s.runCmd = func(context.Context, string, ...string) (string, error) {
		t.Fatal("should not run version")
		return "", nil
	}
	info := s.List()[0]
	if info.Installed || info.Message != "未安装" {
		t.Fatalf("expected not installed: %+v", info)
	}
}

func TestFindProxyDef(t *testing.T) {
	if _, err := findProxyDef("opencode"); err != nil {
		t.Fatal(err)
	}
	if _, err := findProxyDef("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildRunCommand(t *testing.T) {
	s := NewAgentProxyService()
	name, args, err := s.BuildRunCommand("opencode", "hello")
	if err != nil || name != "opencode" || len(args) != 2 || args[0] != "run" || args[1] != "hello" {
		t.Fatalf("opencode: %s %v %v", name, args, err)
	}
	name, args, err = s.BuildRunCommand("claude", "p")
	if err != nil || name != "claude" || args[0] != "-p" {
		t.Fatalf("claude: %s %v %v", name, args, err)
	}
	_, _, err = s.BuildRunCommand("reasonix", "p")
	if err == nil {
		t.Fatal("reasonix should error until non-interactive cmd is documented")
	}
}
