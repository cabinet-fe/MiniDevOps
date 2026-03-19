package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

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
	projectService *service.ProjectService
	cronNotifier   CronNotifier
}

func NewProjectHandler(ps *service.ProjectService, cn CronNotifier) *ProjectHandler {
	return &ProjectHandler{projectService: ps, cronNotifier: cn}
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

// POST /api/v1/projects - create (set created_by)
func (h *ProjectHandler) Create(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required"`
		Description   string `json:"description"`
		RepoURL       string `json:"repo_url" binding:"required"`
		RepoAuthType  string `json:"repo_auth_type"`
		RepoUsername  string `json:"repo_username"`
		RepoPassword  string `json:"repo_password"`
		MaxArtifacts  int    `json:"max_artifacts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.MaxArtifacts == 0 {
		req.MaxArtifacts = 5
	}
	project := &model.Project{
		Name:         req.Name,
		Description:  req.Description,
		RepoURL:      req.RepoURL,
		RepoAuthType: req.RepoAuthType,
		RepoUsername: req.RepoUsername,
		RepoPassword: req.RepoPassword,
		MaxArtifacts: req.MaxArtifacts,
		CreatedBy:    middleware.GetUserID(c),
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
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		RepoURL       *string `json:"repo_url"`
		RepoAuthType  *string `json:"repo_auth_type"`
		RepoUsername  *string `json:"repo_username"`
		RepoPassword  *string `json:"repo_password"`
		MaxArtifacts  *int    `json:"max_artifacts"`
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
	if req.RepoURL != nil {
		project.RepoURL = *req.RepoURL
	}
	if req.RepoAuthType != nil {
		project.RepoAuthType = *req.RepoAuthType
	}
	if req.RepoUsername != nil {
		project.RepoUsername = *req.RepoUsername
	}
	if req.RepoPassword != nil {
		project.RepoPassword = *req.RepoPassword
	}
	if req.MaxArtifacts != nil {
		project.MaxArtifacts = *req.MaxArtifacts
	}
	if err := h.projectService.Update(project); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "更新失败")
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
		Name             string `json:"name" binding:"required"`
		Branch           string `json:"branch"`
		BuildScript      string `json:"build_script"`
		BuildOutputDir   string `json:"build_output_dir"`
		DeployServerID   *uint  `json:"deploy_server_id"`
		DeployPath       string `json:"deploy_path"`
		DeployMethod     string `json:"deploy_method"`
		PostDeployScript string `json:"post_deploy_script"`
		EnvVars          string `json:"env_vars"`
		CronExpression   string `json:"cron_expression"`
		CronEnabled      bool   `json:"cron_enabled"`
		SortOrder        int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.CronEnabled && req.CronExpression != "" {
		if _, err := cron.ParseStandard(req.CronExpression); err != nil {
			pkg.Error(c, http.StatusBadRequest, "Cron 表达式不合法: "+err.Error())
			return
		}
	}
	env := &model.Environment{
		ProjectID:        projectID,
		Name:             req.Name,
		Branch:           req.Branch,
		BuildScript:      req.BuildScript,
		BuildOutputDir:   req.BuildOutputDir,
		DeployServerID:   req.DeployServerID,
		DeployPath:       req.DeployPath,
		DeployMethod:     req.DeployMethod,
		PostDeployScript: req.PostDeployScript,
		EnvVars:          req.EnvVars,
		CronExpression:   req.CronExpression,
		CronEnabled:      req.CronEnabled,
		SortOrder:        req.SortOrder,
	}
	if err := h.projectService.CreateEnvironment(env); err != nil {
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
		Name             *string `json:"name"`
		Branch           *string `json:"branch"`
		BuildScript      *string `json:"build_script"`
		BuildOutputDir   *string `json:"build_output_dir"`
		DeployServerID   *uint   `json:"deploy_server_id"`
		DeployPath       *string `json:"deploy_path"`
		DeployMethod     *string `json:"deploy_method"`
		PostDeployScript *string `json:"post_deploy_script"`
		EnvVars          *string `json:"env_vars"`
		CronExpression   *string `json:"cron_expression"`
		CronEnabled      *bool   `json:"cron_enabled"`
		SortOrder        *int    `json:"sort_order"`
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
	if req.BuildOutputDir != nil {
		env.BuildOutputDir = *req.BuildOutputDir
	}
	if req.DeployServerID != nil {
		env.DeployServerID = req.DeployServerID
	}
	if req.DeployPath != nil {
		env.DeployPath = *req.DeployPath
	}
	if req.DeployMethod != nil {
		env.DeployMethod = *req.DeployMethod
	}
	if req.PostDeployScript != nil {
		env.PostDeployScript = *req.PostDeployScript
	}
	if req.EnvVars != nil {
		env.EnvVars = *req.EnvVars
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
	if err := h.projectService.UpdateEnvironment(env); err != nil {
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
	pkg.Success(c, nil)
}
