package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	authmiddleware "bedrock/internal/auth/middleware"
	"bedrock/internal/pkg"
	"bedrock/internal/rbac"
	rbacmw "bedrock/internal/rbac/middleware"
	"bedrock/internal/rbac/service"
)

type ResourceHandler struct {
	resources *service.ResourceService
	groups    *service.MenuGroupService
	perm      *service.PermissionService
}

func NewResourceHandler(
	resources *service.ResourceService,
	groups *service.MenuGroupService,
	perm *service.PermissionService,
) *ResourceHandler {
	return &ResourceHandler{resources: resources, groups: groups, perm: perm}
}

func (h *ResourceHandler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	res := rg.Group("/rbac/resources", authMW)
	res.GET("", rbacmw.RequirePermission(h.perm, "system_resources:view"), h.List)
	res.GET("/:id", rbacmw.RequirePermission(h.perm, "system_resources:view"), h.Get)
	res.POST("", rbacmw.RequirePermission(h.perm, "system_resources:create"), h.Create)
	res.PUT("/:id", rbacmw.RequirePermission(h.perm, "system_resources:update"), h.Update)
	res.PUT("/:id/icon", rbacmw.RequirePermission(h.perm, "system_resources:update"), h.UpdateIcon)
	res.DELETE("/:id", rbacmw.RequirePermission(h.perm, "system_resources:delete"), h.Delete)

	mg := rg.Group("/menu-groups", authMW)
	mg.GET("", rbacmw.RequirePermission(h.perm, "system_resources:view"), h.ListMenuGroups)
	mg.GET("/:id", rbacmw.RequirePermission(h.perm, "system_resources:view"), h.GetMenuGroup)
	mg.POST("", rbacmw.RequirePermission(h.perm, "system_resources:create"), h.CreateMenuGroup)
	mg.PUT("/:id", rbacmw.RequirePermission(h.perm, "system_resources:update"), h.UpdateMenuGroup)
	mg.DELETE("/:id", rbacmw.RequirePermission(h.perm, "system_resources:delete"), h.DeleteMenuGroup)
}

func (h *ResourceHandler) List(c *gin.Context) {
	filter := service.ListResourcesFilter{
		Keyword: c.Query("keyword"),
		Type:    c.Query("type"),
	}
	if raw := strings.TrimSpace(c.Query("group_id")); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, "group_id 无效")
			return
		}
		gid := uint(id)
		filter.GroupID = &gid
	}
	if raw := strings.TrimSpace(c.Query("enabled")); raw != "" {
		switch strings.ToLower(raw) {
		case "true", "1":
			v := true
			filter.Enabled = &v
		case "false", "0":
			v := false
			filter.Enabled = &v
		default:
			pkg.Error(c, http.StatusBadRequest, "enabled 必须为 true 或 false")
			return
		}
	}
	tree, err := h.resources.ListTree(filter)
	if err != nil {
		if strings.Contains(err.Error(), "type 必须为") {
			pkg.Error(c, http.StatusBadRequest, err.Error())
			return
		}
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
	res, err := h.resources.Create(req, authmiddleware.IsSuperAdmin(c))
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
	res, err := h.resources.Update(id, req, authmiddleware.IsSuperAdmin(c))
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
		if strings.Contains(msg, "32KB") || strings.Contains(msg, "Base64") || strings.Contains(msg, "图标") ||
			strings.Contains(msg, "体积") || strings.Contains(msg, strconv.Itoa(rbac.MaxMenuIconBytes)) {
			pkg.Error(c, http.StatusBadRequest, msg)
			return
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

func (h *ResourceHandler) ListMenuGroups(c *gin.Context) {
	items, err := h.groups.List()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, gin.H{"items": items})
}

func (h *ResourceHandler) GetMenuGroup(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	g, err := h.groups.Get(id)
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "分组不存在")
		return
	}
	pkg.Success(c, g)
}

func (h *ResourceHandler) CreateMenuGroup(c *gin.Context) {
	var req service.CreateMenuGroupInput
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	g, err := h.groups.Create(req)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, g)
}

func (h *ResourceHandler) UpdateMenuGroup(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	var req service.UpdateMenuGroupInput
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	g, err := h.groups.Update(id, req)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, g)
}

func (h *ResourceHandler) DeleteMenuGroup(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	if err := h.groups.Delete(id); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id), err
}
