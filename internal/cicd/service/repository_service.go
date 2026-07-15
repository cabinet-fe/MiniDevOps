package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"bedrock/internal/cicd/model"
	"bedrock/internal/cicd/repository"
	"bedrock/internal/engine"
)

// GitLister abstracts git ls-remote for tests.
type GitLister interface {
	ListBranches(repoURL, authType, username, password string) ([]string, error)
}

type defaultGitLister struct{}

func (defaultGitLister) ListBranches(repoURL, authType, username, password string) ([]string, error) {
	return engine.GitListBranches(repoURL, authType, username, password)
}

type RepositoryService struct {
	repo  *repository.RepositoryRepository
	creds *CredentialService
	git   GitLister
}

func NewRepositoryService(repo *repository.RepositoryRepository, creds *CredentialService) *RepositoryService {
	return &RepositoryService{repo: repo, creds: creds, git: defaultGitLister{}}
}

func (s *RepositoryService) SetGitLister(g GitLister) {
	if g != nil {
		s.git = g
	}
}

type CreateRepositoryInput struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	Tags               string `json:"tags"`
	RepoURL            string `json:"repo_url"`
	DefaultBranch      string `json:"default_branch"`
	AuthType           string `json:"auth_type"`
	CredentialID       *uint  `json:"credential_id"`
	WebhookType        string `json:"webhook_type"`
	WebhookRefPath     string `json:"webhook_ref_path"`
	WebhookCommitPath  string `json:"webhook_commit_path"`
	WebhookMessagePath string `json:"webhook_message_path"`
}

type UpdateRepositoryInput struct {
	Name               *string `json:"name"`
	Description        *string `json:"description"`
	Tags               *string `json:"tags"`
	RepoURL            *string `json:"repo_url"`
	DefaultBranch      *string `json:"default_branch"`
	AuthType           *string `json:"auth_type"`
	CredentialID       *uint   `json:"credential_id"`
	ClearCredential    bool    `json:"clear_credential"`
	WebhookType        *string `json:"webhook_type"`
	WebhookRefPath     *string `json:"webhook_ref_path"`
	WebhookCommitPath  *string `json:"webhook_commit_path"`
	WebhookMessagePath *string `json:"webhook_message_path"`
}

func (s *RepositoryService) Create(createdBy uint, in CreateRepositoryInput, canUseCredential bool) (*model.Repository, error) {
	name := strings.TrimSpace(in.Name)
	url := strings.TrimSpace(in.RepoURL)
	if name == "" || url == "" {
		return nil, errorsNew("名称与仓库 URL 不能为空")
	}
	authType := normalizeRepoAuth(in.AuthType)
	if authType == "credential" {
		if in.CredentialID == nil || *in.CredentialID == 0 {
			return nil, errorsNew("auth_type=credential 时必须提供 credential_id")
		}
		if !canUseCredential {
			return nil, NewForbidden("绑定凭证需要 cicd.credentials:use 权限")
		}
		if _, err := s.creds.Get(*in.CredentialID); err != nil {
			return nil, errorsNew("凭证不存在")
		}
	} else {
		in.CredentialID = nil
	}
	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, err
	}
	branch := strings.TrimSpace(in.DefaultBranch)
	if branch == "" {
		branch = "main"
	}
	whType := strings.TrimSpace(in.WebhookType)
	if whType == "" {
		whType = "auto"
	}
	repo := &model.Repository{
		Name:               name,
		Description:        strings.TrimSpace(in.Description),
		Tags:               strings.TrimSpace(in.Tags),
		RepoURL:            url,
		DefaultBranch:      branch,
		AuthType:           authType,
		CredentialID:       in.CredentialID,
		WebhookSecret:      secret,
		WebhookType:        whType,
		WebhookRefPath:     strings.TrimSpace(in.WebhookRefPath),
		WebhookCommitPath:  strings.TrimSpace(in.WebhookCommitPath),
		WebhookMessagePath: strings.TrimSpace(in.WebhookMessagePath),
		CreatedBy:          createdBy,
	}
	if err := s.repo.Create(repo); err != nil {
		return nil, err
	}
	return publicRepo(repo, false), nil
}

func (s *RepositoryService) Update(id uint, in UpdateRepositoryInput, canUseCredential bool) (*model.Repository, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, NewNotFound("仓库不存在")
	}
	prevCred := existing.CredentialID
	if in.Name != nil {
		existing.Name = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		existing.Description = strings.TrimSpace(*in.Description)
	}
	if in.Tags != nil {
		existing.Tags = strings.TrimSpace(*in.Tags)
	}
	if in.RepoURL != nil {
		existing.RepoURL = strings.TrimSpace(*in.RepoURL)
	}
	if in.DefaultBranch != nil {
		existing.DefaultBranch = strings.TrimSpace(*in.DefaultBranch)
	}
	if in.WebhookType != nil {
		existing.WebhookType = strings.TrimSpace(*in.WebhookType)
	}
	if in.WebhookRefPath != nil {
		existing.WebhookRefPath = strings.TrimSpace(*in.WebhookRefPath)
	}
	if in.WebhookCommitPath != nil {
		existing.WebhookCommitPath = strings.TrimSpace(*in.WebhookCommitPath)
	}
	if in.WebhookMessagePath != nil {
		existing.WebhookMessagePath = strings.TrimSpace(*in.WebhookMessagePath)
	}
	if in.AuthType != nil {
		existing.AuthType = normalizeRepoAuth(*in.AuthType)
	}
	if in.ClearCredential {
		existing.CredentialID = nil
		existing.AuthType = "none"
	} else if in.CredentialID != nil {
		if !credentialIDEqual(prevCred, in.CredentialID) {
			if !canUseCredential {
				return nil, NewForbidden("绑定/修改凭证需要 cicd.credentials:use 权限")
			}
		}
		if *in.CredentialID == 0 {
			existing.CredentialID = nil
		} else {
			if _, err := s.creds.Get(*in.CredentialID); err != nil {
				return nil, errorsNew("凭证不存在")
			}
			existing.CredentialID = in.CredentialID
			existing.AuthType = "credential"
		}
	}
	if existing.AuthType == "credential" && (existing.CredentialID == nil || *existing.CredentialID == 0) {
		return nil, errorsNew("auth_type=credential 时必须提供 credential_id")
	}
	if existing.Name == "" || existing.RepoURL == "" {
		return nil, errorsNew("名称与仓库 URL 不能为空")
	}
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return publicRepo(existing, false), nil
}

func (s *RepositoryService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return NewNotFound("仓库不存在")
	}
	n, err := s.repo.CountJobs(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return NewConflict("该仓库仍被构建任务引用，无法删除")
	}
	return s.repo.Delete(id)
}

func (s *RepositoryService) Get(id uint, revealSecret bool) (*model.Repository, error) {
	repo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, NewNotFound("仓库不存在")
	}
	return publicRepo(repo, revealSecret), nil
}

func (s *RepositoryService) List(page, pageSize int, keyword string) ([]model.Repository, int64, error) {
	items, total, err := s.repo.List(page, pageSize, keyword)
	if err != nil {
		return nil, 0, err
	}
	out := make([]model.Repository, 0, len(items))
	for i := range items {
		out = append(out, *publicRepo(&items[i], false))
	}
	return out, total, nil
}

func (s *RepositoryService) RotateWebhookSecret(id uint) (*model.Repository, error) {
	repo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, NewNotFound("仓库不存在")
	}
	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, err
	}
	repo.WebhookSecret = secret
	if err := s.repo.Update(repo); err != nil {
		return nil, err
	}
	return publicRepo(repo, true), nil
}

func (s *RepositoryService) ListBranches(id uint) ([]string, error) {
	repo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, NewNotFound("仓库不存在")
	}
	authType := "none"
	username, password := "", ""
	if repo.AuthType == "credential" && repo.CredentialID != nil {
		cred, secret, _, err := s.creds.GetDecrypted(*repo.CredentialID)
		if err != nil {
			return nil, err
		}
		username = cred.Username
		password = secret
		authType = "password"
		if cred.Type == "ssh_key" {
			authType = "ssh"
		}
	}
	return s.git.ListBranches(repo.RepoURL, authType, username, password)
}

func (s *RepositoryService) TestFetch(id uint) (map[string]interface{}, error) {
	branches, err := s.ListBranches(id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"ok":           true,
		"branch_count": len(branches),
		"branches":     branches,
	}, nil
}

func publicRepo(repo *model.Repository, revealSecret bool) *model.Repository {
	cp := *repo
	if !revealSecret {
		cp.WebhookSecret = ""
	}
	return &cp
}

func normalizeRepoAuth(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "credential":
		return "credential"
	default:
		return "none"
	}
}

func generateWebhookSecret() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate webhook secret: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func credentialIDEqual(a, b *uint) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
