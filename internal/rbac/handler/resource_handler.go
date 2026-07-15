package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"bedrock/internal/pkg"
	"bedrock/internal/rbac"
	rbacmw "bedrock/internal/rbac/middleware"
	"bedrock/internal/rbac/service"
)

type ResourceHandler struct {
	resources *service.ResourceService
	perm      *service.PermissionService
}

func NewResourceHandler(resources *service.ResourceService, perm *service.PermissionService) *ResourceHandler {
	return &ResourceHandler{resources: resources, perm: perm}
}

func (h *ResourceHandler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	res := rg.Group("/rbac/resources", authMW)
	res.GET("", rbacmw.RequirePermission(h.perm, "system.resources:view"), h.List)
	res.GET("/:id", rbacmw.RequirePermission(h.perm, "system.resources:view"), h.Get)
	res.POST("", rbacmw.RequirePermission(h.perm, "system.resources:create"), h.Create)
	res.PUT("/:id", rbacmw.RequirePermission(h.perm, "system.resources:update"), h.Update)
	res.PUT("/:id/icon", rbacmw.RequirePermission(h.perm, "system.resources:update"), h.UpdateIcon)
	res.DELETE("/:id", rbacmw.RequirePermission(h.perm, "system.resources:delete"), h.Delete)

	// Menu tree for role permission editor (menu-type nodes only).
	menus := rg.Group("/menus", authMW)
	menus.GET("", rbacmw.RequirePermission(h.perm, "system.roles:update"), h.ListMenus)
}

func (h *ResourceHandler) List(c *gin.Context) {
	tree, err := h.resources.ListTree()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, gin.H{"items": tree})
}

func (h *ResourceHandler) ListMenus(c *gin.Context) {
	tree, err := h.resources.ListMenusTree()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, gin.H{"items": tree})
}

func (h *ResourceHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	res, err := h.resources.Get(id)
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "资源不存在")
		return
	}
	pkg.Success(c, res)
}

func (h *ResourceHandler) Create(c *gin.Context) {
	var req service.CreateResourceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	res, err := h.resources.Create(req)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, res)
}

func (h *ResourceHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	var req service.UpdateResourceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	res, err := h.resources.Update(id, req)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, res)
}

func (h *ResourceHandler) UpdateIcon(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	var req struct {
		IconBase64 string `json:"icon_base64" binding:"required"`
		IconMime   string `json:"icon_mime"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	res, err := h.resources.UpdateMenuIcon(id, req.IconBase64, req.IconMime)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "32KB") || strings.Contains(msg, "Base64") || strings.Contains(msg, "图标") {
			pkg.Error(c, http.StatusBadRequest, msg)
			return
		}
		if len(msg) > 0 {
			// size reject must be 400
			if strings.Contains(msg, strconv.Itoa(rbac.MaxMenuIconBytes)) || strings.Contains(msg, "体积") {
				pkg.Error(c, http.StatusBadRequest, msg)
				return
			}
		}
		pkg.Error(c, http.StatusBadRequest, msg)
		return
	}
	pkg.Success(c, res)
}

func (h *ResourceHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	if err := h.resources.Delete(id); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id), err
}
