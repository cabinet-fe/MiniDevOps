package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"minidevops/internal/gitee"
	"minidevops/internal/model"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	client     *model.Client
	giteeClient *gitee.Client
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(client *model.Client) *ProjectHandler {
	return &ProjectHandler{
		client:     client,
		giteeClient: gitee.NewClient(""), // 默认无token
	}
}

// ProjectRequest 项目请求
type ProjectRequest struct {
	Name     string `json:"name"`
	RepoURL  string `json:"repo_url"`
	Branch   string `json:"branch"`
	BuildCmd string `json:"build_cmd"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	RepoURL     string     `json:"repo_url"`
	Branch      string     `json:"branch"`
	BuildCmd    string     `json:"build_cmd"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastBuildAt *time.Time `json:"last_build_at,omitempty"`
}

// CreateProject godoc
// @Summary 创建新项目
// @Description 创建新的构建项目
// @Tags 项目
// @Accept json
// @Produce json
// @Param project body ProjectRequest true "项目信息"
// @Success 201 {object} ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/projects [post]
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var req ProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的请求格式",
		})
	}

	// 验证请求数据
	if req.Name == "" || req.RepoURL == "" || req.BuildCmd == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "项目名称、仓库URL和构建命令不能为空",
		})
	}

	// 获取当前用户ID
	userID := c.Locals("userID").(int)
	user, err := h.client.User.Get(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "获取用户信息失败",
		})
	}

	// 使用用户的码云Token验证仓库
	if user.GiteeToken != "" {
		h.giteeClient = gitee.NewClient(user.GiteeToken)
	}

	// 校验仓库是否存在
	_, err = h.giteeClient.GetRepoFromURL(req.RepoURL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "仓库验证失败: " + err.Error(),
		})
	}

	// 如果分支为空，设为默认值
	if req.Branch == "" {
		req.Branch = "master"
	}

	// 创建项目
	now := time.Now()
	project, err := h.client.Project.
		Create().
		SetName(req.Name).
		SetRepoURL(req.RepoURL).
		SetBranch(req.Branch).
		SetBuildCmd(req.BuildCmd).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		SetOwner(user).
		Save(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "创建项目失败",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(ProjectResponse{
		ID:        project.ID,
		Name:      project.Name,
		RepoURL:   project.RepoURL,
		Branch:    project.Branch,
		BuildCmd:  project.BuildCmd,
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	})
}

// GetProjects godoc
// @Summary 获取项目列表
// @Description 获取当前用户的所有项目
// @Tags 项目
// @Produce json
// @Success 200 {array} ProjectResponse
// @Failure 401 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/projects [get]
func (h *ProjectHandler) GetProjects(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int)

	// 查询用户的所有项目
	projects, err := h.client.Project.
		Query().
		Where(model.HasOwnerWith(model.UserIDEQ(userID))).
		All(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "获取项目列表失败",
		})
	}

	// 转换为响应格式
	response := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		response[i] = ProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			RepoURL:     project.RepoURL,
			Branch:      project.Branch,
			BuildCmd:    project.BuildCmd,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
			LastBuildAt: project.LastBuildAt,
		}
	}

	return c.JSON(response)
}

// GetProject godoc
// @Summary 获取项目详情
// @Description 获取指定项目的详细信息
// @Tags 项目
// @Produce json
// @Param id path int true "项目ID"
// @Success 200 {object} ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的项目ID",
		})
	}

	userID := c.Locals("userID").(int)

	// 查询项目
	project, err := h.client.Project.
		Query().
		Where(
			model.ProjectIDEQ(id),
			model.HasOwnerWith(model.UserIDEQ(userID)),
		).
		Only(c.Context())

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "项目不存在或无权访问",
		})
	}

	return c.JSON(ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		RepoURL:     project.RepoURL,
		Branch:      project.Branch,
		BuildCmd:    project.BuildCmd,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		LastBuildAt: project.LastBuildAt,
	})
}

// UpdateProject godoc
// @Summary 更新项目
// @Description 更新指定项目的信息
// @Tags 项目
// @Accept json
// @Produce json
// @Param id path int true "项目ID"
// @Param project body ProjectRequest true "项目更新信息"
// @Success 200 {object} ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的项目ID",
		})
	}

	var req ProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的请求格式",
		})
	}

	userID := c.Locals("userID").(int)

	// 查询项目
	project, err := h.client.Project.
		Query().
		Where(
			model.ProjectIDEQ(id),
			model.HasOwnerWith(model.UserIDEQ(userID)),
		).
		Only(c.Context())

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "项目不存在或无权访问",
		})
	}

	// 如果仓库URL发生变更，需要验证新仓库
	if req.RepoURL != "" && req.RepoURL != project.RepoURL {
		// 获取用户的码云Token
		user, err := h.client.User.Get(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error: "获取用户信息失败",
			})
		}

		// 使用用户的码云Token验证仓库
		if user.GiteeToken != "" {
			h.giteeClient = gitee.NewClient(user.GiteeToken)
		}

		// 校验仓库是否存在
		_, err = h.giteeClient.GetRepoFromURL(req.RepoURL)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error: "仓库验证失败: " + err.Error(),
			})
		}
	}

	// 准备更新
	update := project.Update()
	if req.Name != "" {
		update = update.SetName(req.Name)
	}
	if req.RepoURL != "" {
		update = update.SetRepoURL(req.RepoURL)
	}
	if req.Branch != "" {
		update = update.SetBranch(req.Branch)
	}
	if req.BuildCmd != "" {
		update = update.SetBuildCmd(req.BuildCmd)
	}

	// 更新项目
	project, err = update.
		SetUpdatedAt(time.Now()).
		Save(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "更新项目失败",
		})
	}

	return c.JSON(ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		RepoURL:     project.RepoURL,
		Branch:      project.Branch,
		BuildCmd:    project.BuildCmd,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		LastBuildAt: project.LastBuildAt,
	})
}

// DeleteProject godoc
// @Summary 删除项目
// @Description 删除指定的项目
// @Tags 项目
// @Produce json
// @Param id path int true "项目ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的项目ID",
		})
	}

	userID := c.Locals("userID").(int)

	// 查询项目
	exists, err := h.client.Project.
		Query().
		Where(
			model.ProjectIDEQ(id),
			model.HasOwnerWith(model.UserIDEQ(userID)),
		).
		Exist(c.Context())

	if err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "项目不存在或无权访问",
		})
	}

	// 删除项目
	err = h.client.Project.DeleteOneID(id).Exec(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "删除项目失败",
		})
	}

	return c.JSON(SuccessResponse{
		Message: "项目已成功删除",
	})
}