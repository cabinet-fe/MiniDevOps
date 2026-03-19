package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"buildflow/internal/config"
	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"
)

type ProjectService struct {
	repo      *repository.ProjectRepository
	envRepo   *repository.EnvironmentRepository
	buildRepo *repository.BuildRepository
}

func NewProjectService(repo *repository.ProjectRepository, envRepo *repository.EnvironmentRepository, buildRepo *repository.BuildRepository) *ProjectService {
	return &ProjectService{repo: repo, envRepo: envRepo, buildRepo: buildRepo}
}

func generateWebhookSecret() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *ProjectService) Create(project *model.Project) error {
	secret, err := generateWebhookSecret()
	if err != nil {
		return err
	}
	project.WebhookSecret = secret
	if project.RepoPassword != "" {
		enc, err := pkg.Encrypt(project.RepoPassword)
		if err != nil {
			return err
		}
		project.RepoPassword = enc
	}
	return s.repo.Create(project)
}

func (s *ProjectService) GetByID(id uint) (*model.Project, error) {
	return s.repo.FindByID(id)
}

func (s *ProjectService) List(page, pageSize int, role string, createdBy uint) ([]model.Project, int64, error) {
	var filter *uint
	if role == "dev" {
		filter = &createdBy
	}
	return s.repo.List(page, pageSize, filter)
}

func (s *ProjectService) Update(project *model.Project) error {
	existing, err := s.repo.FindByID(project.ID)
	if err != nil {
		return err
	}
	if project.RepoPassword != "" && project.RepoPassword != existing.RepoPassword {
		enc, err := pkg.Encrypt(project.RepoPassword)
		if err != nil {
			return err
		}
		project.RepoPassword = enc
	} else {
		project.RepoPassword = existing.RepoPassword
	}
	return s.repo.Update(project)
}

func (s *ProjectService) Delete(id uint) error {
	if err := s.envRepo.DeleteByProjectID(id); err != nil {
		return err
	}
	if err := s.buildRepo.DeleteByProjectID(id); err != nil {
		return err
	}
	// Clean workspace and artifact directories
	if config.C != nil {
		workspaceDir := filepath.Join(config.C.Build.WorkspaceDir, fmt.Sprintf("project-%d", id))
		_ = os.RemoveAll(workspaceDir)
		artifactDir := filepath.Join(config.C.Build.ArtifactDir, fmt.Sprintf("project-%d", id))
		_ = os.RemoveAll(artifactDir)
	}
	return s.repo.Delete(id)
}

// ProjectExport is a safe representation for export (no sensitive fields).
type ProjectExport struct {
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	RepoURL       string              `json:"repo_url"`
	RepoAuthType  string              `json:"repo_auth_type"`
	RepoUsername  string              `json:"repo_username"`
	MaxArtifacts  int                 `json:"max_artifacts"`
	Environments  []EnvironmentExport `json:"environments"`
}

// EnvironmentExport excludes sensitive fields.
type EnvironmentExport struct {
	Name             string `json:"name"`
	Branch           string `json:"branch"`
	BuildScript      string `json:"build_script"`
	BuildOutputDir   string `json:"build_output_dir"`
	DeployPath       string `json:"deploy_path"`
	DeployMethod     string `json:"deploy_method"`
	PostDeployScript string `json:"post_deploy_script"`
	EnvVars          string `json:"env_vars"`
	CronExpression   string `json:"cron_expression"`
	CronEnabled      bool   `json:"cron_enabled"`
	SortOrder        int    `json:"sort_order"`
}

func (s *ProjectService) Export(id uint) ([]byte, error) {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	exp := ProjectExport{
		Name:          project.Name,
		Description:   project.Description,
		RepoURL:       project.RepoURL,
		RepoAuthType:  project.RepoAuthType,
		RepoUsername:  project.RepoUsername,
		MaxArtifacts:  project.MaxArtifacts,
		Environments:  make([]EnvironmentExport, 0, len(project.Environments)),
	}
	for _, e := range project.Environments {
		exp.Environments = append(exp.Environments, EnvironmentExport{
			Name:             e.Name,
			Branch:           e.Branch,
			BuildScript:      e.BuildScript,
			BuildOutputDir:   e.BuildOutputDir,
			DeployPath:       e.DeployPath,
			DeployMethod:     e.DeployMethod,
			PostDeployScript: e.PostDeployScript,
			EnvVars:          e.EnvVars,
			CronExpression:   e.CronExpression,
			CronEnabled:      e.CronEnabled,
			SortOrder:        e.SortOrder,
		})
	}
	return json.MarshalIndent(exp, "", "  ")
}

func (s *ProjectService) Import(data []byte, createdBy uint) (*model.Project, error) {
	var exp ProjectExport
	if err := json.Unmarshal(data, &exp); err != nil {
		return nil, err
	}
	baseName := exp.Name
	name := baseName
	suffix := 0
	for {
		_, err := s.repo.FindByName(name)
		if err != nil {
			break
		}
		suffix++
		name = fmt.Sprintf("%s_%d", baseName, suffix)
	}
	project := &model.Project{
		Name:          name,
		Description:   exp.Description,
		RepoURL:       exp.RepoURL,
		RepoAuthType:  exp.RepoAuthType,
		RepoUsername:  exp.RepoUsername,
		MaxArtifacts:  exp.MaxArtifacts,
		CreatedBy:     createdBy,
	}
	if err := s.repo.Create(project); err != nil {
		return nil, err
	}
	for _, ee := range exp.Environments {
		env := &model.Environment{
			ProjectID:        project.ID,
			Name:             ee.Name,
			Branch:           ee.Branch,
			BuildScript:      ee.BuildScript,
			BuildOutputDir:   ee.BuildOutputDir,
			DeployPath:       ee.DeployPath,
			DeployMethod:     ee.DeployMethod,
			PostDeployScript: ee.PostDeployScript,
			EnvVars:          ee.EnvVars,
			CronExpression:   ee.CronExpression,
			CronEnabled:      ee.CronEnabled,
			SortOrder:        ee.SortOrder,
		}
		if err := s.envRepo.Create(env); err != nil {
			return nil, err
		}
	}
	return s.repo.FindByID(project.ID)
}

func (s *ProjectService) ListEnvironments(projectID uint) ([]model.Environment, error) {
	return s.envRepo.ListByProjectID(projectID)
}

func (s *ProjectService) CreateEnvironment(env *model.Environment) error {
	return s.envRepo.Create(env)
}

func (s *ProjectService) UpdateEnvironment(env *model.Environment) error {
	existing, err := s.envRepo.FindByID(env.ID)
	if err != nil {
		return err
	}
	if existing.ProjectID != env.ProjectID {
		return errors.New("环境不属于该项目")
	}
	return s.envRepo.Update(env)
}

func (s *ProjectService) DeleteEnvironment(id, projectID uint) error {
	env, err := s.envRepo.FindByID(id)
	if err != nil {
		return err
	}
	if env.ProjectID != projectID {
		return errors.New("环境不属于该项目")
	}
	return s.envRepo.Delete(id)
}
