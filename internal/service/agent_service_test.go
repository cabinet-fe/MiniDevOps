package service

import (
	"testing"

	"buildflow/internal/model"
)

func TestIsValidProxyKey(t *testing.T) {
	if !IsValidProxyKey("opencode") || !IsValidProxyKey("claude") || !IsValidProxyKey("reasonix") {
		t.Fatal("expected known keys valid")
	}
	if IsValidProxyKey("unknown") || IsValidProxyKey("") {
		t.Fatal("expected unknown keys invalid")
	}
}

func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name    string
		agent   *model.Agent
		wantErr bool
	}{
		{"ok", &model.Agent{Name: "review", ProxyKey: "opencode"}, false},
		{"trim name", &model.Agent{Name: "  x  ", ProxyKey: "claude"}, false},
		{"empty name", &model.Agent{Name: "  ", ProxyKey: "opencode"}, true},
		{"bad proxy", &model.Agent{Name: "a", ProxyKey: "foo"}, true},
		{"reasonix", &model.Agent{Name: "r", ProxyKey: "reasonix"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgent(tt.agent)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAgent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.agent.Name == "" {
				t.Fatal("expected name trimmed non-empty")
			}
		})
	}
}
