package engine

import (
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		repoURL  string
		wantName string
	}{
		{"https://github.com/user/repo.git", "github"},
		{"https://gitlab.com/user/repo.git", "gitlab"},
		{"https://gitlab.mycompany.com/user/repo.git", "gitlab"},
		{"https://gitee.com/user/repo.git", "gitee"},
		{"https://gitea.example.com/user/repo.git", "gitea"},
		{"https://custom.example.com/user/repo.git", "generic"},
	}
	for _, tt := range tests {
		t.Run(tt.repoURL, func(t *testing.T) {
			p := DetectPlatform(tt.repoURL)
			if p.Name() != tt.wantName {
				t.Errorf("DetectPlatform(%q) = %q, want %q", tt.repoURL, p.Name(), tt.wantName)
			}
		})
	}
}

func TestBuildAuthURL_Token(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		username string
		token    string
		wantUser string // expected user part in URL
	}{
		{
			name:     "github without username",
			repoURL:  "https://github.com/user/repo.git",
			username: "",
			token:    "ghp_abc123",
			wantUser: "x-access-token",
		},
		{
			name:     "gitlab without username",
			repoURL:  "https://gitlab.com/user/repo.git",
			username: "",
			token:    "glpat-abc123",
			wantUser: "oauth2",
		},
		{
			name:     "gitee with username",
			repoURL:  "https://gitee.com/user/repo.git",
			username: "myuser",
			token:    "abc123",
			wantUser: "myuser",
		},
		{
			name:     "gitee without username falls back",
			repoURL:  "https://gitee.com/user/repo.git",
			username: "",
			token:    "abc123",
			wantUser: "oauth2",
		},
		{
			name:     "custom username overrides default",
			repoURL:  "https://github.com/user/repo.git",
			username: "deploy-token",
			token:    "abc123",
			wantUser: "deploy-token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAuthURL(tt.repoURL, "token", tt.username, tt.token)
			if result == "" {
				t.Fatal("buildAuthURL returned empty string")
			}
			// Check that the URL contains the expected username
			expected := tt.wantUser + ":" + tt.token + "@"
			if !contains(result, expected) {
				t.Errorf("buildAuthURL() = %q, want to contain %q", result, expected)
			}
		})
	}
}

func TestBuildAuthURL_Password(t *testing.T) {
	result := buildAuthURL("https://github.com/user/repo.git", "password", "myuser", "mypass")
	expected := "myuser:mypass@"
	if !contains(result, expected) {
		t.Errorf("buildAuthURL() = %q, want to contain %q", result, expected)
	}
}

func TestBuildAuthURL_None(t *testing.T) {
	raw := "https://github.com/user/repo.git"
	result := buildAuthURL(raw, "none", "", "")
	if result != raw {
		t.Errorf("buildAuthURL(none) = %q, want %q", result, raw)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
