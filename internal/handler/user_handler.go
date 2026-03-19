package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"buildflow/internal/middleware"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(us *service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

// GET /api/v1/users - paginated list
func (h *UserHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPage(c)
	users, total, err := h.userService.List(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Paginated(c, users, total, page, pageSize)
}

// POST /api/v1/users - create user (admin)
func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Email       string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Role == "" {
		req.Role = "dev"
	}
	user, err := h.userService.Create(req.Username, req.Password, req.DisplayName, req.Role, req.Email)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, user)
}

// GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	user, err := h.userService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "用户不存在")
		return
	}
	pkg.Success(c, user)
}

// PUT /api/v1/users/:id - update user (role, display_name, email, is_active)
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	user, err := h.userService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "用户不存在")
		return
	}
	var req struct {
		Role        *string `json:"role"`
		DisplayName *string `json:"display_name"`
		Email       *string `json:"email"`
		IsActive    *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if err := h.userService.Update(user); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	pkg.Success(c, user)
}

// DELETE /api/v1/users/:id - prevent self-deletion
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	userID := uint(id)
	currentUserID := middleware.GetUserID(c)
	if currentUserID == userID {
		pkg.Error(c, http.StatusForbidden, "不能删除自己的账号")
		return
	}
	if err := h.userService.Delete(userID); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}
