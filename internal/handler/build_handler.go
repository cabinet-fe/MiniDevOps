package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"buildflow/internal/config"
	"buildflow/internal/middleware"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

// BuildScheduler submits builds for execution. Used to avoid circular dependency.
type BuildScheduler interface {
	Submit(buildID uint)
}

type BuildHandler struct {
	buildService *service.BuildService
	scheduler    BuildScheduler
}

func NewBuildHandler(bs *service.BuildService, scheduler BuildScheduler) *BuildHandler {
	return &BuildHandler{buildService: bs, scheduler: scheduler}
}

// GET /api/v1/projects/:id/builds - list builds for project
func (h *BuildHandler) ListByProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	page, pageSize := pkg.GetPage(c)
	var envID *uint
	if e := c.Query("environment_id"); e != "" {
		if parsed, err := strconv.ParseUint(e, 10, 32); err == nil {
			id := uint(parsed)
			envID = &id
		}
	}
	builds, total, err := h.buildService.ListByProject(uint(projectID), envID, page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Paginated(c, builds, total, page, pageSize)
}

// POST /api/v1/projects/:id/builds - trigger build (accepts environment_id in JSON body)
func (h *BuildHandler) TriggerBuild(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	var req struct {
		EnvironmentID uint `json:"environment_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.EnvironmentID == 0 {
		pkg.Error(c, http.StatusBadRequest, "environment_id 必填")
		return
	}
	userID := middleware.GetUserID(c)
	build, err := h.buildService.TriggerBuild(uint(projectID), req.EnvironmentID, userID, "manual", "", "")
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if h.scheduler != nil {
		h.scheduler.Submit(build.ID)
	}
	pkg.Created(c, build)
}

// GET /api/v1/builds/:id - build detail
func (h *BuildHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	build, err := h.buildService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "构建不存在")
		return
	}
	pkg.Success(c, build)
}

// POST /api/v1/builds/:id/cancel - cancel build
func (h *BuildHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.buildService.Cancel(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "取消失败")
		return
	}
	pkg.Success(c, nil)
}

// POST /api/v1/builds/:id/deploy - manual deploy
func (h *BuildHandler) Deploy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	build, err := h.buildService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "构建不存在")
		return
	}
	if build.Status != "success" {
		pkg.Error(c, http.StatusBadRequest, "只有成功的构建才能部署")
		return
	}
	if h.scheduler != nil {
		h.scheduler.Submit(build.ID)
	}
	pkg.Success(c, gin.H{"message": "部署已提交"})
}

// GET /api/v1/builds/:id/artifact - download artifact file
func (h *BuildHandler) DownloadArtifact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	build, err := h.buildService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "构建不存在")
		return
	}
	if build.ArtifactPath == "" {
		pkg.Error(c, http.StatusNotFound, "该构建没有产物")
		return
	}
	path := build.ArtifactPath
	if !filepath.IsAbs(path) && config.C != nil {
		path = filepath.Join(config.C.Build.ArtifactDir, path)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		pkg.Error(c, http.StatusNotFound, "产物文件不存在")
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(build.ArtifactPath))
	c.File(path)
}

// POST /api/v1/builds/:id/rollback - rollback
func (h *BuildHandler) Rollback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	build, err := h.buildService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "构建不存在")
		return
	}
	if build.Status != "success" {
		pkg.Error(c, http.StatusBadRequest, "只能回滚到成功的构建")
		return
	}
	if h.scheduler != nil {
		h.scheduler.Submit(build.ID)
	}
	pkg.Success(c, gin.H{"message": "回滚已提交"})
}

// GET /api/v1/dashboard/stats
func (h *BuildHandler) DashboardStats(c *gin.Context) {
	stats, err := h.buildService.GetDashboardStats()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, stats)
}

// GET /api/v1/dashboard/active-builds
func (h *BuildHandler) DashboardActiveBuilds(c *gin.Context) {
	builds, err := h.buildService.GetActiveBuildsList()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, builds)
}

// GET /api/v1/dashboard/recent-builds
func (h *BuildHandler) DashboardRecentBuilds(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	builds, err := h.buildService.GetRecentBuilds(limit)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, builds)
}

// GET /api/v1/dashboard/trend - build trend for last N days
func (h *BuildHandler) DashboardTrend(c *gin.Context) {
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 30 {
			days = parsed
		}
	}
	trend, err := h.buildService.GetBuildTrend(days)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, trend)
}
