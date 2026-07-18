package service

import "testing"

func TestNpmPackageFromTemplate(t *testing.T) {
	got := npmPackageFromTemplate(npmCLIInstallLike("@anthropic-ai/claude-code"))
	if got != "@anthropic-ai/claude-code" {
		t.Fatalf("got %q", got)
	}
	if npmPackageFromTemplate(`curl -fsSL "$base/install.sh" | sh`) != "" {
		t.Fatal("expected empty for non-npm template")
	}
}

func npmCLIInstallLike(pkg string) string {
	return `version="{{version}}"; base="{{base_url}}"; reg=""; [ -n "$base" ] && reg="--registry $base"; npm install -g ` + pkg + `${version:+@$version} $reg`
}

func TestIsNewerCLIVersion(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"2.0.0", "1.9.9", true},
		{"1.2.3", "1.2.3", false},
		{"1.2.3", "1.2.4", false},
		{"1.2.4-beta", "1.2.4", false},
		{"1.2.4", "1.2.4-beta", true},
		{"v2.1.0", "2.0.9", true},
		{"", "1.0.0", false},
		{"1.0.0", "", false},
	}
	for _, tc := range cases {
		if got := isNewerCLIVersion(tc.latest, tc.current); got != tc.want {
			t.Fatalf("isNewer(%q,%q)=%v want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestNormalizeCLIVersion(t *testing.T) {
	if got := normalizeCLIVersion("claude version 1.2.3"); got != "1.2.3" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeCLIVersion("/usr/local/bin/claude"); got != "" {
		t.Fatalf("path should be empty, got %q", got)
	}
}
