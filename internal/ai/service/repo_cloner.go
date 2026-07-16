package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cicdmodel "bedrock/internal/cicd/model"
)

// RepositoryFinder loads repository config for agent workspace cloning.
type RepositoryFinder interface {
	FindByID(id uint) (*cicdmodel.Repository, error)
}

// SimpleRepoCloner shallow-clones public/local-accessible repos for agent context.
// Credentialed clones reuse the repository URL as configured; secrets are not logged.
type SimpleRepoCloner struct {
	repos RepositoryFinder
}

func NewSimpleRepoCloner(repos RepositoryFinder) *SimpleRepoCloner {
	return &SimpleRepoCloner{repos: repos}
}

func (c *SimpleRepoCloner) CloneForAgent(ctx context.Context, repositoryID uint, destDir string) error {
	if c == nil || c.repos == nil {
		return fmt.Errorf("repository cloner not configured")
	}
	repo, err := c.repos.FindByID(repositoryID)
	if err != nil {
		return err
	}
	url := strings.TrimSpace(repo.RepoURL)
	if url == "" {
		return fmt.Errorf("repository URL empty")
	}
	branch := strings.TrimSpace(repo.DefaultBranch)
	if branch == "" {
		branch = "main"
	}
	if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
		return err
	}
	args := []string{"clone", "--depth", "1", "--branch", branch, url, destDir}
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}
