package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"buildflow/internal/engine"
	"buildflow/internal/middleware"
	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

// CronNotifier is called when an environment's cron settings change.
type CronNotifier interface {
	Add(env model.Environment) error
	Remove(envID uint)
}

type ProjectHandler struct {
	projectService    *service.ProjectService
	credentialService *service.CredentialService
	cronNotifier      CronNotifier
}

func NewProjectHandler(ps *service.ProjectService, cs *service.CredentialService, cn CronNotifier) *ProjectHandler {
	return &ProjectHandler{projectService: ps, credentialService: cs, cronNotifier: cn}
}

// GET /api/v1/projects - list (pass role and user_id for filtering)
func (h *ProjectHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPage(c)
	role := middleware.GetRole(c)
	userID := middleware.GetUserID(c)
	projects, total, err := h.projectService.List(page, pageSize, role, userID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Paginated(c, projects, total, page, pageSize)
}

// GET /api/v1/environments - paginated environments across projects (optional project_id, name)
func (h *ProjectHandler) ListEnvironmentsGlobal(c *gin.Context) {
	page, pageSize := pkg.GetPage(c)
	role := middleware.GetRole(c)
	userID := middleware.GetUserID(c)
	var projectID *uint
	if pid := strings.TrimSpace(c.Query("project_id")); pid != "" {
		if id, err := strconv.ParseUint(pid, 10, 32); err == nil {
			u := uint(id)
			projectID = &u
		}
	}
	name := c.Query("name")
	items, total, err := h.projectService.ListEnvironmentsGlobal(page, pageSize, projectID, name, role, userID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Paginated(c, items, total, page, pageSize)
}

// POST /api/v1/projects - create (set created_by)
func (h *ProjectHandler) Create(c *gin.Context) {
	var req struct {
		Name               string `json:"name" binding:"required"`
		Description        string `json:"description"`
		Tags               string `json:"tags"`
		RepoURL            string `json:"repo_url" binding:"required"`
		RepoAuthType       string `json:"repo_auth_type"`
		CredentialID       *uint  `json:"credential_id"`
		MaxArtifacts       int    `json:"max_artifacts"`
		ArtifactFormat     string `json:"artifact_format"`
		WebhookType        string `json:"webhook_type"`
		WebhookRefPath     string `json:"webhook_ref_path"`
		WebhookCommitPath  string `json:"webhook_commit_path"`
		WebhookMessagePath string `json:"webhook_message_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.MaxArtifacts == 0 {
		req.MaxArtifacts = 5
	}
	req.ArtifactFormat = normalizeArtifactFormat(req.ArtifactFormat)
	project := &model.Project{
		Name:               req.Name,
		Description:        req.Description,
		Tags:               req.Tags,
		RepoURL:            req.RepoURL,
		RepoAuthType:       req.RepoAuthType,
		CredentialID:       req.CredentialID,
		MaxArtifacts:       req.MaxArtifacts,
		ArtifactFormat:     req.ArtifactFormat,
		WebhookType:        req.WebhookType,
		WebhookRefPath:     req.WebhookRefPath,
		WebhookCommitPath:  req.WebhookCommitPath,
		WebhookMessagePath: req.WebhookMessagePath,
		CreatedBy:          middleware.GetUserID(c),
	}
	if project.RepoAuthType == "credential" && project.CredentialID != nil {
		allowed, err := h.credentialService.CanUseCredential(*project.CredentialID, middleware.GetUserID(c), middleware.GetRole(c))
		if err != nil {
			pkg.Error(c, http.StatusInternalServerError, "校验凭证失败")
			return
		}
		if !allowed {
			pkg.Error(c, http.StatusForbidden, "无权使用该凭证")
			return
		}
	}
	if err := h.projectService.Create(project); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, project)
}

// GET /api/v1/projects/:id - detail with environments
func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	project, err := h.projectService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	pkg.Success(c, project)
}

// PUT /api/v1/projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	project, err := h.projectService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	var req struct {
		Name               *string `json:"name"`
		Description        *string `json:"description"`
		Tags               *string `json:"tags"`
		RepoURL            *string `json:"repo_url"`
		RepoAuthType       *string `json:"repo_auth_type"`
		CredentialID       *uint   `json:"credential_id"`
		MaxArtifacts       *int    `json:"max_artifacts"`
		ArtifactFormat     *string `json:"artifact_format"`
		WebhookType        *string `json:"webhook_type"`
		WebhookRefPath     *string `json:"webhook_ref_path"`
		WebhookCommitPath  *string `json:"webhook_commit_path"`
		WebhookMessagePath *string `json:"webhook_message_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Tags != nil {
		project.Tags = *req.Tags
	}
	if req.RepoURL != nil {
		project.RepoURL = *req.RepoURL
	}
	if req.RepoAuthType != nil {
		project.RepoAuthType = *req.RepoAuthType
	}
	if req.CredentialID != nil {
		project.CredentialID = req.CredentialID
	}
	if req.MaxArtifacts != nil {
		project.MaxArtifacts = *req.MaxArtifacts
	}
	if req.ArtifactFormat != nil {
		project.ArtifactFormat = normalizeArtifactFormat(*req.ArtifactFormat)
	}
	if req.WebhookType != nil {
		project.WebhookType = *req.WebhookType
	}
	if req.WebhookRefPath != nil {
		project.WebhookRefPath = *req.WebhookRefPath
	}
	if req.WebhookCommitPath != nil {
		project.WebhookCommitPath = *req.WebhookCommitPath
	}
	if req.WebhookMessagePath != nil {
		project.WebhookMessagePath = *req.WebhookMessagePath
	}
	if project.RepoAuthType == "credential" && project.CredentialID != nil {
		allowed, err := h.credentialService.CanUseCredential(*project.CredentialID, middleware.GetUserID(c), middleware.GetRole(c))
		if err != nil {
			pkg.Error(c, http.StatusInternalServerError, "校验凭证失败")
			return
		}
		if !allowed {
			pkg.Error(c, http.StatusForbidden, "无权使用该凭证")
			return
		}
	}
	if err := h.projectService.Update(project); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, project)
}

// DELETE /api/v1/projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	project, err := h.projectService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	if h.cronNotifier != nil {
		for _, env := range project.Environments {
			h.cronNotifier.Remove(env.ID)
		}
	}
	if err := h.projectService.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

// GET /api/v1/projects/:id/export
func (h *ProjectHandler) Export(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	data, err := h.projectService.Export(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=project-export.json")
	c.Data(http.StatusOK, "application/json", data)
}

// POST /api/v1/projects/import
func (h *ProjectHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "请上传文件")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "读取文件失败")
		return
	}
	createdBy := middleware.GetUserID(c)
	project, err := h.projectService.Import(data, createdBy)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, project)
}

// GET /api/v1/projects/:id/envs
func (h *ProjectHandler) ListEnvironments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	envs, err := h.projectService.ListEnvironments(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, envs)
}

// POST /api/v1/projects/:id/envs
func (h *ProjectHandler) CreateEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	projectID := uint(id)
	var req struct {
		Name            string                 `json:"name" binding:"required"`
		Branch          string                 `json:"branch"`
		BuildScript     string                 `json:"build_script"`
		BuildScriptType string                 `json:"build_script_type"`
		BuildOutputDir  string                 `json:"build_output_dir"`
		Distributions   []model.Distribution   `json:"distributions"`
		CachePaths      string                 `json:"cache_paths"`
		CronExpression  string                 `json:"cron_expression"`
		CronEnabled     bool                   `json:"cron_enabled"`
		SortOrder       int                    `json:"sort_order"`
		VarGroupIDs     []uint                 `json:"var_group_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.BuildScriptType == "" {
		req.BuildScriptType = "bash"
	}
	if req.CronEnabled && req.CronExpression != "" {
		if _, err := cron.ParseStandard(req.CronExpression); err != nil {
			pkg.Error(c, http.StatusBadRequest, "Cron 表达式不合法: "+err.Error())
			return
		}
	}
	env := &model.Environment{
		ProjectID:       projectID,
		Name:            req.Name,
		Branch:          req.Branch,
		BuildScript:     req.BuildScript,
		BuildScriptType: req.BuildScriptType,
		BuildOutputDir:  req.BuildOutputDir,
		CachePaths:      req.CachePaths,
		CronExpression:  req.CronExpression,
		CronEnabled:     req.CronEnabled,
		SortOrder:       req.SortOrder,
	}
	if err := h.projectService.CreateEnvironment(env, req.VarGroupIDs, req.Distributions); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if h.cronNotifier != nil && env.CronEnabled {
		_ = h.cronNotifier.Add(*env)
	}
	pkg.Created(c, env)
}

// PUT /api/v1/projects/:id/envs/:envId
func (h *ProjectHandler) UpdateEnvironment(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	envID, err := strconv.ParseUint(c.Param("envId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	envs, err := h.projectService.ListEnvironments(uint(projectID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	var env *model.Environment
	for i := range envs {
		if envs[i].ID == uint(envID) {
			env = &envs[i]
			break
		}
	}
	if env == nil {
		pkg.Error(c, http.StatusNotFound, "环境不存在")
		return
	}
	var req struct {
		Name            *string                `json:"name"`
		Branch          *string                `json:"branch"`
		BuildScript     *string                `json:"build_script"`
		BuildScriptType *string                `json:"build_script_type"`
		BuildOutputDir  *string                `json:"build_output_dir"`
		Distributions   []model.Distribution   `json:"distributions"`
		CachePaths      *string                `json:"cache_paths"`
		CronExpression  *string                `json:"cron_expression"`
		CronEnabled     *bool                  `json:"cron_enabled"`
		SortOrder       *int                   `json:"sort_order"`
		VarGroupIDs     []uint                 `json:"var_group_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Name != nil {
		env.Name = *req.Name
	}
	if req.Branch != nil {
		env.Branch = *req.Branch
	}
	if req.BuildScript != nil {
		env.BuildScript = *req.BuildScript
	}
	if req.BuildScriptType != nil {
		env.BuildScriptType = *req.BuildScriptType
	}
	if req.BuildOutputDir != nil {
		env.BuildOutputDir = *req.BuildOutputDir
	}
	if req.CachePaths != nil {
		env.CachePaths = *req.CachePaths
	}
	if req.CronExpression != nil {
		env.CronExpression = *req.CronExpression
	}
	if req.CronEnabled != nil {
		env.CronEnabled = *req.CronEnabled
	}
	// Validate cron expression if enabled
	if env.CronEnabled && env.CronExpression != "" {
		if _, err := cron.ParseStandard(env.CronExpression); err != nil {
			pkg.Error(c, http.StatusBadRequest, "Cron 表达式不合法: "+err.Error())
			return
		}
	}
	if req.SortOrder != nil {
		env.SortOrder = *req.SortOrder
	}
	syncDistributions := req.Distributions != nil
	if err := h.projectService.UpdateEnvironment(env, req.VarGroupIDs, req.VarGroupIDs != nil, req.Distributions, syncDistributions); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	// Notify cron scheduler of changes
	if h.cronNotifier != nil {
		if env.CronEnabled && env.CronExpression != "" {
			_ = h.cronNotifier.Add(*env)
		} else {
			h.cronNotifier.Remove(env.ID)
		}
	}
	pkg.Success(c, env)
}

// DELETE /api/v1/projects/:id/envs/:envId
func (h *ProjectHandler) DeleteEnvironment(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	envID, err := strconv.ParseUint(c.Param("envId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.projectService.DeleteEnvironment(uint(envID), uint(projectID)); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if h.cronNotifier != nil {
		h.cronNotifier.Remove(uint(envID))
	}
	pkg.Success(c, nil)
}

// GET /api/v1/projects/:id/envs/:envId/vars
func (h *ProjectHandler) ListEnvVars(c *gin.Context) {
	projectID, envID, ok := parseProjectEnvIDs(c)
	if !ok {
		return
	}
	items, err := h.projectService.ListEnvVars(projectID, envID)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, items)
}

// POST /api/v1/projects/:id/envs/:envId/vars
func (h *ProjectHandler) CreateEnvVar(c *gin.Context) {
	projectID, envID, ok := parseProjectEnvIDs(c)
	if !ok {
		return
	}
	var req struct {
		Key      string `json:"key" binding:"required"`
		Value    string `json:"value"`
		IsSecret bool   `json:"is_secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.projectService.CreateEnvVar(projectID, envID, req.Key, req.Value, req.IsSecret); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.projectService.ListEnvVars(projectID, envID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Created(c, items)
}

// PUT /api/v1/projects/:id/envs/:envId/vars/:varId
func (h *ProjectHandler) UpdateEnvVar(c *gin.Context) {
	projectID, envID, ok := parseProjectEnvIDs(c)
	if !ok {
		return
	}
	varID, err := strconv.ParseUint(c.Param("varId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	var req struct {
		Key       string `json:"key" binding:"required"`
		Value     string `json:"value"`
		IsSecret  bool   `json:"is_secret"`
		KeepValue bool   `json:"keep_value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	item, err := h.projectService.UpdateEnvVar(projectID, envID, uint(varID), req.Key, req.Value, req.IsSecret, req.KeepValue)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, item)
}

func normalizeArtifactFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "zip":
		return "zip"
	default:
		return "gzip"
	}
}

// DELETE /api/v1/projects/:id/envs/:envId/vars/:varId
func (h *ProjectHandler) DeleteEnvVar(c *gin.Context) {
	projectID, envID, ok := parseProjectEnvIDs(c)
	if !ok {
		return
	}
	varID, err := strconv.ParseUint(c.Param("varId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.projectService.DeleteEnvVar(projectID, envID, uint(varID)); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

// GET /api/v1/var-groups
func (h *ProjectHandler) ListVarGroups(c *gin.Context) {
	groups, err := h.projectService.ListVarGroups()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, groups)
}

// POST /api/v1/var-groups
func (h *ProjectHandler) CreateVarGroup(c *gin.Context) {
	group, ok := bindVarGroupPayload(c)
	if !ok {
		return
	}
	if err := h.projectService.CreateVarGroup(group); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, group)
}

// PUT /api/v1/var-groups/:groupId
func (h *ProjectHandler) UpdateVarGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("groupId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	group, ok := bindVarGroupPayload(c)
	if !ok {
		return
	}
	group.ID = uint(groupID)
	if err := h.projectService.UpdateVarGroup(group); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, group)
}

// DELETE /api/v1/var-groups/:groupId
func (h *ProjectHandler) DeleteVarGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("groupId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.projectService.DeleteVarGroup(uint(groupID)); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

// GET /api/v1/projects/:id/branches - list remote branches
func (h *ProjectHandler) ListBranches(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	project, err := h.projectService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	authType, username, password, err := h.projectService.ResolveRepoAuth(project)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "获取凭证失败: "+err.Error())
		return
	}
	branches, err := engine.GitListBranches(project.RepoURL, authType, username, password)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "获取分支列表失败: "+err.Error())
		return
	}
	pkg.Success(c, branches)
}

func parseProjectEnvIDs(c *gin.Context) (uint, uint, bool) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return 0, 0, false
	}
	envID, err := strconv.ParseUint(c.Param("envId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return 0, 0, false
	}
	return uint(projectID), uint(envID), true
}

func bindVarGroupPayload(c *gin.Context) (*model.VarGroup, bool) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Items       []struct {
			ID        uint   `json:"id"`
			Key       string `json:"key" binding:"required"`
			Value     string `json:"value"`
			IsSecret  bool   `json:"is_secret"`
			KeepValue bool   `json:"keep_value"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return nil, false
	}
	group := &model.VarGroup{
		Name:        req.Name,
		Description: req.Description,
		Items:       make([]model.VarGroupItem, 0, len(req.Items)),
	}
	for _, item := range req.Items {
		group.Items = append(group.Items, model.VarGroupItem{
			ID:       item.ID,
			Key:      item.Key,
			Value:    item.Value,
			IsSecret: item.IsSecret,
			HasValue: item.KeepValue,
		})
	}
	return group, true
}
