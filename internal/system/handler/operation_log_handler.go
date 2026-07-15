package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"bedrock/internal/pkg"
	rbacmw "bedrock/internal/rbac/middleware"
	rbacservice "bedrock/internal/rbac/service"
	"bedrock/internal/system/repository"
	"bedrock/internal/system/service"
)

type OperationLogHandler struct {
	audit *service.AuditService
	perm  *rbacservice.PermissionService
}

func NewOperationLogHandler(audit *service.AuditService, perm *rbacservice.PermissionService) *OperationLogHandler {
	return &OperationLogHandler{audit: audit, perm: perm}
}

func (h *OperationLogHandler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/operation-logs", authMW)
	g.GET("", rbacmw.RequirePermission(h.perm, "system.operation_logs:view"), h.List)
}

func (h *OperationLogHandler) List(c *gin.Context) {
	page := pkg.ParsePage(c)
	f := repository.OperationLogFilters{
		Page:         page.Page,
		PageSize:     page.PageSize,
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
	}
	if uid := c.Query("user_id"); uid != "" {
		if v, err := strconv.ParseUint(uid, 10, 64); err == nil {
			id := uint(v)
			f.UserID = &id
		}
	}
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			f.From = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			end := t.Add(24*time.Hour - time.Nanosecond)
			f.To = &end
		}
	}
	items, total, err := h.audit.List(f)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.PageSuccess(c, items, total, page)
}
