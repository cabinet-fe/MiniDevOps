package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	authmiddleware "bedrock/internal/auth/middleware"
	"bedrock/internal/dashboard/model"
	"bedrock/internal/dashboard/service"
	"bedrock/internal/pkg"
	rbacmw "bedrock/internal/rbac/middleware"
	rbacservice "bedrock/internal/rbac/service"
)

type DashboardHandler struct {
	svc  *service.DashboardService
	perm *rbacservice.PermissionService
}

func NewDashboardHandler(svc *service.DashboardService, perm *rbacservice.PermissionService) *DashboardHandler {
	return &DashboardHandler{svc: svc, perm: perm}
}

func (h *DashboardHandler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/dashboard", authMW, rbacmw.RequirePermission(h.perm, "dashboard:view"))
	g.GET("/layout", h.GetLayout)
	g.PUT("/layout", h.PutLayout)
	g.GET("/build-summary", rbacmw.RequirePermission(h.perm, "cicd_build_runs:view"), h.BuildSummary)
	g.GET("/agent-run-summary", rbacmw.RequirePermission(h.perm, "ai_runs:view"), h.AgentRunSummary)
	g.GET("/system-info", rbacmw.RequirePermission(h.perm, "dashboard:system_info"), h.SystemInfo)
	g.GET("/system-status", rbacmw.RequirePermission(h.perm, "dashboard:system_status"), h.SystemStatus)
}

func (h *DashboardHandler) GetLayout(c *gin.Context) {
	perms, ok := h.permissions(c)
	if !ok {
		return
	}
	layout, err := h.svc.GetLayout(authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c), perms)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "读取仪表盘布局失败")
		return
	}
	pkg.Success(c, layout)
}

func (h *DashboardHandler) PutLayout(c *gin.Context) {
	var req model.LayoutResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效仪表盘布局")
		return
	}
	perms, ok := h.permissions(c)
	if !ok {
		return
	}
	layout, err := h.svc.PutLayout(authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c), perms, req.Cards)
	if errors.Is(err, service.ErrUnauthorizedCard) {
		pkg.Error(c, http.StatusForbidden, "不能添加无权限卡片")
		return
	}
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, layout)
}

func (h *DashboardHandler) BuildSummary(c *gin.Context) {
	result, err := h.svc.BuildSummary()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "读取构建摘要失败")
		return
	}
	pkg.Success(c, result)
}

func (h *DashboardHandler) AgentRunSummary(c *gin.Context) {
	result, err := h.svc.AgentRunSummary()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "读取智能体运行摘要失败")
		return
	}
	pkg.Success(c, result)
}

func (h *DashboardHandler) SystemInfo(c *gin.Context) {
	result, err := h.svc.SystemInfo()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "读取系统信息失败")
		return
	}
	pkg.Success(c, result)
}

func (h *DashboardHandler) SystemStatus(c *gin.Context) {
	result, err := h.svc.SystemStatus()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "读取系统状态失败")
		return
	}
	pkg.Success(c, result)
}

func (h *DashboardHandler) permissions(c *gin.Context) ([]string, bool) {
	perms, err := h.perm.ResolvePermissions(authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "权限校验失败")
		return nil, false
	}
	return perms, true
}
