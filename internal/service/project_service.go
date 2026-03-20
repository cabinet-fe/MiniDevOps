package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"buildflow/internal/config"
	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"
)

type ProjectService struct {
	repo         *repository.ProjectRepository
	envRepo      *repository.EnvironmentRepository
	buildRepo    *repository.BuildRepository
	envVarRepo   *repository.EnvVarRepository
	varGroupRepo *repository.VarGroupRepository
}

func NewProjectService(
	repo *repository.ProjectRepository,
	envRepo *repository.EnvironmentRepository,
	buildRepo *repository.BuildRepository,
	envVarRepo *repository.EnvVarRepository,
	varGroupRepo *repository.VarGroupRepository,
) *ProjectService {
	return &ProjectService{
		repo:         repo,
		envRepo:      envRepo,
		buildRepo:    buildRepo,
		envVarRepo:   envVarRepo,
		varGroupRepo: varGroupRepo,
	}
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
	if project.WebhookType == "" {
		project.WebhookType = "auto"
	}
	project.ArtifactFormat = normalizeProjectArtifactFormat(project.ArtifactFormat)
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
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.populateProjectEnvironmentGroupIDs(project); err != nil {
		return nil, err
	}
	return project, nil
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
	if project.WebhookType == "" {
		project.WebhookType = existing.WebhookType
	}
	project.ArtifactFormat = normalizeProjectArtifactFormat(project.ArtifactFormat)
	return s.repo.Update(project)
}

func (s *ProjectService) Delete(id uint) error {
	if err := s.envVarRepo.DeleteByProjectID(id); err != nil {
		return err
	}
	if err := s.varGroupRepo.DeleteLinksByProjectID(id); err != nil {
		return err
	}
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
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	GroupName          string              `json:"group_name"`
	Tags               string              `json:"tags"`
	RepoURL            string              `json:"repo_url"`
	RepoAuthType       string              `json:"repo_auth_type"`
	RepoUsername       string              `json:"repo_username"`
	MaxArtifacts       int                 `json:"max_artifacts"`
	ArtifactFormat     string              `json:"artifact_format"`
	WebhookType        string              `json:"webhook_type"`
	WebhookRefPath     string              `json:"webhook_ref_path"`
	WebhookCommitPath  string              `json:"webhook_commit_path"`
	WebhookMessagePath string              `json:"webhook_message_path"`
	Environments       []EnvironmentExport `json:"environments"`
}

// EnvironmentExport excludes sensitive fields.
type EnvironmentExport struct {
	Name             string         `json:"name"`
	Branch           string         `json:"branch"`
	BuildScript      string         `json:"build_script"`
	BuildScriptType  string         `json:"build_script_type"`
	BuildOutputDir   string         `json:"build_output_dir"`
	DeployPath       string         `json:"deploy_path"`
	DeployMethod     string         `json:"deploy_method"`
	PostDeployScript string         `json:"post_deploy_script"`
	CachePaths       string         `json:"cache_paths"`
	CronExpression   string         `json:"cron_expression"`
	CronEnabled      bool           `json:"cron_enabled"`
	SortOrder        int            `json:"sort_order"`
	VarGroupNames    []string       `json:"var_group_names,omitempty"`
	EnvVars          []EnvVarExport `json:"env_vars,omitempty"`
}

type EnvVarExport struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

func (s *ProjectService) Export(id uint) ([]byte, error) {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	exp := ProjectExport{
		Name:               project.Name,
		Description:        project.Description,
		GroupName:          project.GroupName,
		Tags:               project.Tags,
		RepoURL:            project.RepoURL,
		RepoAuthType:       project.RepoAuthType,
		RepoUsername:       project.RepoUsername,
		MaxArtifacts:       project.MaxArtifacts,
		ArtifactFormat:     normalizeProjectArtifactFormat(project.ArtifactFormat),
		WebhookType:        project.WebhookType,
		WebhookRefPath:     project.WebhookRefPath,
		WebhookCommitPath:  project.WebhookCommitPath,
		WebhookMessagePath: project.WebhookMessagePath,
		Environments:       make([]EnvironmentExport, 0, len(project.Environments)),
	}
	for _, e := range project.Environments {
		varGroupNames, _ := s.listEnvironmentGroupNames(e.ID)
		envVars, _ := s.ListEnvVars(e.ProjectID, e.ID)
		exportVars := make([]EnvVarExport, 0, len(envVars))
		for _, envVar := range envVars {
			value := envVar.Value
			if envVar.IsSecret {
				value = ""
			}
			exportVars = append(exportVars, EnvVarExport{
				Key:      envVar.Key,
				Value:    value,
				IsSecret: envVar.IsSecret,
			})
		}
		exp.Environments = append(exp.Environments, EnvironmentExport{
			Name:             e.Name,
			Branch:           e.Branch,
			BuildScript:      e.BuildScript,
			BuildScriptType:  e.BuildScriptType,
			BuildOutputDir:   e.BuildOutputDir,
			DeployPath:       e.DeployPath,
			DeployMethod:     e.DeployMethod,
			PostDeployScript: e.PostDeployScript,
			CachePaths:       e.CachePaths,
			CronExpression:   e.CronExpression,
			CronEnabled:      e.CronEnabled,
			SortOrder:        e.SortOrder,
			VarGroupNames:    varGroupNames,
			EnvVars:          exportVars,
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
		Name:               name,
		Description:        exp.Description,
		GroupName:          exp.GroupName,
		Tags:               exp.Tags,
		RepoURL:            exp.RepoURL,
		RepoAuthType:       exp.RepoAuthType,
		RepoUsername:       exp.RepoUsername,
		MaxArtifacts:       exp.MaxArtifacts,
		ArtifactFormat:     normalizeProjectArtifactFormat(exp.ArtifactFormat),
		WebhookType:        exp.WebhookType,
		WebhookRefPath:     exp.WebhookRefPath,
		WebhookCommitPath:  exp.WebhookCommitPath,
		WebhookMessagePath: exp.WebhookMessagePath,
		CreatedBy:          createdBy,
	}
	if err := s.Create(project); err != nil {
		return nil, err
	}
	for _, ee := range exp.Environments {
		env := &model.Environment{
			ProjectID:        project.ID,
			Name:             ee.Name,
			Branch:           ee.Branch,
			BuildScript:      ee.BuildScript,
			BuildScriptType:  ee.BuildScriptType,
			BuildOutputDir:   ee.BuildOutputDir,
			DeployPath:       ee.DeployPath,
			DeployMethod:     ee.DeployMethod,
			PostDeployScript: ee.PostDeployScript,
			CachePaths:       ee.CachePaths,
			CronExpression:   ee.CronExpression,
			CronEnabled:      ee.CronEnabled,
			SortOrder:        ee.SortOrder,
		}
		if err := s.CreateEnvironment(env, nil); err != nil {
			return nil, err
		}
		for _, exported := range ee.EnvVars {
			if err := s.CreateEnvVar(project.ID, env.ID, exported.Key, exported.Value, exported.IsSecret); err != nil {
				return nil, err
			}
		}
		if len(ee.VarGroupNames) > 0 {
			groupIDs := make([]uint, 0, len(ee.VarGroupNames))
			for _, groupName := range ee.VarGroupNames {
				group, err := s.varGroupRepo.FindByName(groupName)
				if err == nil {
					groupIDs = append(groupIDs, group.ID)
				}
			}
			if err := s.varGroupRepo.SetEnvironmentVarGroupIDs(env.ID, groupIDs); err != nil {
				return nil, err
			}
		}
	}
	return s.repo.FindByID(project.ID)
}

func normalizeProjectArtifactFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "zip":
		return "zip"
	default:
		return "gzip"
	}
}

func (s *ProjectService) ListEnvironments(projectID uint) ([]model.Environment, error) {
	envs, err := s.envRepo.ListByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	if err := s.populateEnvironmentGroupIDs(envs); err != nil {
		return nil, err
	}
	return envs, nil
}

func (s *ProjectService) CreateEnvironment(env *model.Environment, varGroupIDs []uint) error {
	if err := s.envRepo.Create(env); err != nil {
		return err
	}
	if varGroupIDs != nil {
		if err := s.varGroupRepo.SetEnvironmentVarGroupIDs(env.ID, uniqueUintSlice(varGroupIDs)); err != nil {
			return err
		}
		env.VarGroupIDs = uniqueUintSlice(varGroupIDs)
	}
	return nil
}

func (s *ProjectService) UpdateEnvironment(env *model.Environment, varGroupIDs []uint, syncVarGroups bool) error {
	existing, err := s.envRepo.FindByID(env.ID)
	if err != nil {
		return err
	}
	if existing.ProjectID != env.ProjectID {
		return errors.New("环境不属于该项目")
	}
	if err := s.envRepo.Update(env); err != nil {
		return err
	}
	if syncVarGroups {
		groupIDs := uniqueUintSlice(varGroupIDs)
		if err := s.varGroupRepo.SetEnvironmentVarGroupIDs(env.ID, groupIDs); err != nil {
			return err
		}
		env.VarGroupIDs = groupIDs
	}
	return nil
}

func (s *ProjectService) DeleteEnvironment(id, projectID uint) error {
	env, err := s.envRepo.FindByID(id)
	if err != nil {
		return err
	}
	if env.ProjectID != projectID {
		return errors.New("环境不属于该项目")
	}
	if err := s.envVarRepo.DeleteByEnvironmentID(id); err != nil {
		return err
	}
	if err := s.varGroupRepo.DeleteEnvironmentLinks(id); err != nil {
		return err
	}
	return s.envRepo.Delete(id)
}

func (s *ProjectService) ListEnvVars(projectID, envID uint) ([]model.EnvVar, error) {
	if _, err := s.getProjectEnvironment(projectID, envID); err != nil {
		return nil, err
	}
	vars, err := s.envVarRepo.ListByEnvironmentID(envID)
	if err != nil {
		return nil, err
	}
	items := make([]model.EnvVar, 0, len(vars))
	for _, item := range vars {
		items = append(items, maskEnvVar(item))
	}
	return items, nil
}

func (s *ProjectService) CreateEnvVar(projectID, envID uint, key, value string, isSecret bool) error {
	if _, err := s.getProjectEnvironment(projectID, envID); err != nil {
		return err
	}
	stored, err := encryptSecretValue(value, isSecret)
	if err != nil {
		return err
	}
	return s.envVarRepo.Create(&model.EnvVar{
		EnvironmentID: envID,
		Key:           strings.TrimSpace(key),
		Value:         stored,
		IsSecret:      isSecret,
	})
}

func (s *ProjectService) UpdateEnvVar(projectID, envID, varID uint, key, value string, isSecret, keepValue bool) (*model.EnvVar, error) {
	if _, err := s.getProjectEnvironment(projectID, envID); err != nil {
		return nil, err
	}
	envVar, err := s.envVarRepo.FindByID(varID)
	if err != nil {
		return nil, err
	}
	if envVar.EnvironmentID != envID {
		return nil, errors.New("变量不属于该环境")
	}
	envVar.Key = strings.TrimSpace(key)
	if keepValue {
		if !isSecret && envVar.IsSecret {
			plain, err := pkg.Decrypt(envVar.Value)
			if err != nil {
				return nil, err
			}
			envVar.Value = plain
		}
	} else {
		stored, err := encryptSecretValue(value, isSecret)
		if err != nil {
			return nil, err
		}
		envVar.Value = stored
	}
	envVar.IsSecret = isSecret
	if err := s.envVarRepo.Update(envVar); err != nil {
		return nil, err
	}
	masked := maskEnvVar(*envVar)
	return &masked, nil
}

func (s *ProjectService) DeleteEnvVar(projectID, envID, varID uint) error {
	if _, err := s.getProjectEnvironment(projectID, envID); err != nil {
		return err
	}
	envVar, err := s.envVarRepo.FindByID(varID)
	if err != nil {
		return err
	}
	if envVar.EnvironmentID != envID {
		return errors.New("变量不属于该环境")
	}
	return s.envVarRepo.Delete(varID)
}

func (s *ProjectService) ListVarGroups() ([]model.VarGroup, error) {
	groups, err := s.varGroupRepo.List()
	if err != nil {
		return nil, err
	}
	items := make([]model.VarGroup, 0, len(groups))
	for _, group := range groups {
		items = append(items, maskVarGroup(group))
	}
	return items, nil
}

func (s *ProjectService) CreateVarGroup(group *model.VarGroup) error {
	prepared, err := prepareVarGroupForStorage(group)
	if err != nil {
		return err
	}
	return s.varGroupRepo.Create(prepared)
}

func (s *ProjectService) UpdateVarGroup(group *model.VarGroup) error {
	existing, err := s.varGroupRepo.FindByID(group.ID)
	if err != nil {
		return err
	}
	existing.Name = strings.TrimSpace(group.Name)
	existing.Description = strings.TrimSpace(group.Description)
	if err := s.varGroupRepo.Update(existing); err != nil {
		return err
	}
	items := make([]model.VarGroupItem, 0, len(group.Items))
	for _, item := range group.Items {
		stored, err := encryptSecretValue(item.Value, item.IsSecret)
		if item.HasValue && item.IsSecret && item.Value == "" {
			stored = item.Value
		}
		if err != nil {
			return err
		}
		if item.HasValue && item.IsSecret && item.Value == "" {
			if matched := findExistingGroupItem(existing.Items, item.ID); matched != nil {
				stored = matched.Value
			}
		}
		items = append(items, model.VarGroupItem{
			Key:      strings.TrimSpace(item.Key),
			Value:    stored,
			IsSecret: item.IsSecret,
		})
	}
	return s.varGroupRepo.ReplaceItems(group.ID, items)
}

func (s *ProjectService) DeleteVarGroup(id uint) error {
	return s.varGroupRepo.Delete(id)
}

func (s *ProjectService) ResolveEnvironmentVars(environmentID uint) ([]string, error) {
	groupItems, err := s.varGroupRepo.ListItemsByEnvironmentID(environmentID)
	if err != nil {
		return nil, err
	}
	envVars, err := s.envVarRepo.ListByEnvironmentID(environmentID)
	if err != nil {
		return nil, err
	}
	merged := make(map[string]string)
	order := make([]string, 0, len(groupItems)+len(envVars))
	for _, item := range groupItems {
		value, err := decryptSecretValue(item.Value, item.IsSecret)
		if err != nil {
			return nil, err
		}
		if _, exists := merged[item.Key]; !exists {
			order = append(order, item.Key)
		}
		merged[item.Key] = value
	}
	for _, item := range envVars {
		value, err := decryptSecretValue(item.Value, item.IsSecret)
		if err != nil {
			return nil, err
		}
		if _, exists := merged[item.Key]; !exists {
			order = append(order, item.Key)
		}
		merged[item.Key] = value
	}
	result := make([]string, 0, len(order))
	for _, key := range order {
		result = append(result, key+"="+merged[key])
	}
	return result, nil
}

func (s *ProjectService) GetEnvironmentVarGroupIDs(environmentID uint) ([]uint, error) {
	return s.varGroupRepo.ListEnvironmentVarGroupIDs(environmentID)
}

func (s *ProjectService) populateProjectEnvironmentGroupIDs(project *model.Project) error {
	if len(project.Environments) == 0 {
		return nil
	}
	return s.populateEnvironmentGroupIDs(project.Environments)
}

func (s *ProjectService) populateEnvironmentGroupIDs(envs []model.Environment) error {
	for i := range envs {
		groupIDs, err := s.varGroupRepo.ListEnvironmentVarGroupIDs(envs[i].ID)
		if err != nil {
			return err
		}
		envs[i].VarGroupIDs = groupIDs
	}
	return nil
}

func (s *ProjectService) getProjectEnvironment(projectID, envID uint) (*model.Environment, error) {
	env, err := s.envRepo.FindByID(envID)
	if err != nil {
		return nil, err
	}
	if env.ProjectID != projectID {
		return nil, errors.New("环境不属于该项目")
	}
	return env, nil
}

func (s *ProjectService) listEnvironmentGroupNames(environmentID uint) ([]string, error) {
	groupIDs, err := s.varGroupRepo.ListEnvironmentVarGroupIDs(environmentID)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		group, err := s.varGroupRepo.FindByID(groupID)
		if err == nil {
			names = append(names, group.Name)
		}
	}
	return names, nil
}

func encryptSecretValue(value string, isSecret bool) (string, error) {
	if !isSecret {
		return value, nil
	}
	return pkg.Encrypt(value)
}

func decryptSecretValue(value string, isSecret bool) (string, error) {
	if !isSecret {
		return value, nil
	}
	return pkg.Decrypt(value)
}

func maskEnvVar(item model.EnvVar) model.EnvVar {
	if item.IsSecret {
		item.HasValue = item.Value != ""
		item.Value = "***"
		item.Masked = true
	}
	return item
}

func maskVarGroup(group model.VarGroup) model.VarGroup {
	items := make([]model.VarGroupItem, 0, len(group.Items))
	for _, item := range group.Items {
		if item.IsSecret {
			item.HasValue = item.Value != ""
			item.Value = "***"
			item.Masked = true
		}
		items = append(items, item)
	}
	group.Items = items
	return group
}

func prepareVarGroupForStorage(group *model.VarGroup) (*model.VarGroup, error) {
	prepared := &model.VarGroup{
		Name:        strings.TrimSpace(group.Name),
		Description: strings.TrimSpace(group.Description),
		Items:       make([]model.VarGroupItem, 0, len(group.Items)),
	}
	for _, item := range group.Items {
		stored, err := encryptSecretValue(item.Value, item.IsSecret)
		if err != nil {
			return nil, err
		}
		prepared.Items = append(prepared.Items, model.VarGroupItem{
			Key:      strings.TrimSpace(item.Key),
			Value:    stored,
			IsSecret: item.IsSecret,
		})
	}
	return prepared, nil
}

func uniqueUintSlice(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func findExistingGroupItem(items []model.VarGroupItem, id uint) *model.VarGroupItem {
	for i := range items {
		if items[i].ID == id {
			return &items[i]
		}
	}
	return nil
}
