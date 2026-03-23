package handler

import (
	"errors"
	"net/http"
	"strconv"

	"buildflow/internal/middleware"
	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CredentialHandler struct {
	credentialService *service.CredentialService
}

func NewCredentialHandler(cs *service.CredentialService) *CredentialHandler {
	return &CredentialHandler{credentialService: cs}
}

// GET /api/v1/credentials
func (h *CredentialHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)
	items, err := h.credentialService.ListByUser(userID, role)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, items)
}

// POST /api/v1/credentials
func (h *CredentialHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Username    string `json:"username"`
		Password    string `json:"password" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	credential := &model.Credential{
		Name:        req.Name,
		Type:        req.Type,
		Username:    req.Username,
		Password:    req.Password,
		Description: req.Description,
		CreatedBy:   middleware.GetUserID(c),
	}
	if err := h.credentialService.Create(credential); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.credentialService.GetByID(credential.ID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "创建成功但读取失败")
		return
	}
	pkg.Created(c, item)
}

// GET /api/v1/credentials/:id
func (h *CredentialHandler) GetByID(c *gin.Context) {
	id, ok := parseCredentialID(c)
	if !ok {
		return
	}
	item, err := h.credentialService.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.Error(c, http.StatusNotFound, "凭证不存在")
			return
		}
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	if !canReadCredential(c, item) {
		pkg.Error(c, http.StatusForbidden, "forbidden")
		return
	}
	pkg.Success(c, item)
}

// PUT /api/v1/credentials/:id
func (h *CredentialHandler) Update(c *gin.Context) {
	id, ok := parseCredentialID(c)
	if !ok {
		return
	}
	existing, err := h.credentialService.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.Error(c, http.StatusNotFound, "凭证不存在")
			return
		}
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	if middleware.GetUserID(c) != existing.CreatedBy {
		pkg.Error(c, http.StatusForbidden, "forbidden")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Type        *string `json:"type"`
		Username    *string `json:"username"`
		Password    *string `json:"password"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	updated := &model.Credential{
		ID:          id,
		Name:        existing.Name,
		Type:        existing.Type,
		Username:    existing.Username,
		Description: existing.Description,
		Password:    "",
	}
	if req.Name != nil {
		updated.Name = *req.Name
	}
	if req.Type != nil {
		updated.Type = *req.Type
	}
	if req.Username != nil {
		updated.Username = *req.Username
	}
	if req.Password != nil {
		updated.Password = *req.Password
	}
	if req.Description != nil {
		updated.Description = *req.Description
	}

	if err := h.credentialService.Update(updated); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.credentialService.GetByID(id)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "更新成功但读取失败")
		return
	}
	pkg.Success(c, item)
}

// DELETE /api/v1/credentials/:id
func (h *CredentialHandler) Delete(c *gin.Context) {
	id, ok := parseCredentialID(c)
	if !ok {
		return
	}
	item, err := h.credentialService.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.Error(c, http.StatusNotFound, "凭证不存在")
			return
		}
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	if middleware.GetUserID(c) != item.CreatedBy {
		pkg.Error(c, http.StatusForbidden, "forbidden")
		return
	}
	if err := h.credentialService.Delete(id); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

// GET /api/v1/credentials/select
func (h *CredentialHandler) ListForSelect(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)
	items, err := h.credentialService.ListForSelect(userID, role)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, items)
}

func parseCredentialID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return 0, false
	}
	return uint(id), true
}

func canReadCredential(c *gin.Context, credential *model.Credential) bool {
	role := middleware.GetRole(c)
	if role == "admin" {
		return true
	}
	return middleware.GetUserID(c) == credential.CreatedBy
}
