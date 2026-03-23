package engine

import (
	"net/url"
	"strings"
)

// GitPlatform defines per-platform Git HTTP authentication behavior.
type GitPlatform interface {
	// Name returns the platform identifier (e.g. "github", "gitlab").
	Name() string
	// BuildTokenAuthURL injects token credentials into a parsed URL.
	// username may be empty; implementations decide the fallback.
	BuildTokenAuthURL(u *url.URL, username, token string) *url.URL
}

// --- GitHub ---

type gitHubPlatform struct{}

func (gitHubPlatform) Name() string { return "github" }

func (gitHubPlatform) BuildTokenAuthURL(u *url.URL, username, token string) *url.URL {
	// GitHub PATs work with "x-access-token" as the conventional username.
	user := username
	if user == "" {
		user = "x-access-token"
	}
	u.User = url.UserPassword(user, token)
	return u
}

// --- GitLab ---

type gitLabPlatform struct{}

func (gitLabPlatform) Name() string { return "gitlab" }

func (gitLabPlatform) BuildTokenAuthURL(u *url.URL, username, token string) *url.URL {
	user := username
	if user == "" {
		user = "oauth2"
	}
	u.User = url.UserPassword(user, token)
	return u
}

// --- Gitee ---

type giteePlatform struct{}

func (giteePlatform) Name() string { return "gitee" }

func (giteePlatform) BuildTokenAuthURL(u *url.URL, username, token string) *url.URL {
	// Gitee requires the real username for token auth.
	// If username is empty, fall back to using the token as a password-only
	// credential which Gitee also accepts via basic auth.
	user := username
	if user == "" {
		user = "oauth2"
	}
	u.User = url.UserPassword(user, token)
	return u
}

// --- Gitea ---

type giteaPlatform struct{}

func (giteaPlatform) Name() string { return "gitea" }

func (giteaPlatform) BuildTokenAuthURL(u *url.URL, username, token string) *url.URL {
	user := username
	if user == "" {
		user = "oauth2"
	}
	u.User = url.UserPassword(user, token)
	return u
}

// --- Generic fallback ---

type genericPlatform struct{}

func (genericPlatform) Name() string { return "generic" }

func (genericPlatform) BuildTokenAuthURL(u *url.URL, username, token string) *url.URL {
	user := username
	if user == "" {
		user = "oauth2"
	}
	u.User = url.UserPassword(user, token)
	return u
}

// platformRegistry maps hostname keywords to platform implementations.
var platformRegistry = []struct {
	keyword  string
	platform GitPlatform
}{
	{"github.com", gitHubPlatform{}},
	{"github", gitHubPlatform{}},
	{"gitlab.com", gitLabPlatform{}},
	{"gitlab", gitLabPlatform{}},
	{"gitee.com", giteePlatform{}},
	{"gitee", giteePlatform{}},
	{"gitea", giteaPlatform{}},
}

var defaultPlatform GitPlatform = genericPlatform{}

// DetectPlatform identifies the Git hosting platform from a repository URL.
// It inspects the hostname for known keywords. Self-hosted instances are
// detected by substring match (e.g. "gitlab.mycompany.com" → GitLab).
func DetectPlatform(repoURL string) GitPlatform {
	u, err := url.Parse(repoURL)
	if err != nil {
		return defaultPlatform
	}
	host := strings.ToLower(u.Hostname())
	for _, entry := range platformRegistry {
		if strings.Contains(host, entry.keyword) {
			return entry.platform
		}
	}
	return defaultPlatform
}
