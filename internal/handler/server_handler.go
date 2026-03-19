package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"buildflow/internal/middleware"
	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

type ServerHandler struct {
	serverService *service.ServerService
}

func NewServerHandler(ss *service.ServerService) *ServerHandler {
	return &ServerHandler{serverService: ss}
}

// GET /api/v1/servers - list with tag filter, pass role for credentials visibility
func (h *ServerHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPage(c)
	tag := c.Query("tag")
	role := middleware.GetRole(c)
	servers, total, err := h.serverService.List(page, pageSize, tag, role)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Paginated(c, servers, total, page, pageSize)
}

// POST /api/v1/servers - create (set created_by from middleware)
func (h *ServerHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Host        string `json:"host" binding:"required"`
		Port        int    `json:"port"`
		Username    string `json:"username" binding:"required"`
		AuthType    string `json:"auth_type" binding:"required"`
		Password    string `json:"password"`
		PrivateKey  string `json:"private_key"`
		Description string `json:"description"`
		Tags        string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Port == 0 {
		req.Port = 22
	}
	server := &model.Server{
		Name:        req.Name,
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		AuthType:    req.AuthType,
		Password:    req.Password,
		PrivateKey:  req.PrivateKey,
		Description: req.Description,
		Tags:        req.Tags,
		CreatedBy:   middleware.GetUserID(c),
	}
	if err := h.serverService.Create(server); err != nil {
		pkg.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.Created(c, server)
}

// GET /api/v1/servers/:id
func (h *ServerHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	role := middleware.GetRole(c)
	server, err := h.serverService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "服务器不存在")
		return
	}
	if role == "dev" {
		server.Password = ""
		server.PrivateKey = ""
	}
	pkg.Success(c, server)
}

// PUT /api/v1/servers/:id
func (h *ServerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	server, err := h.serverService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "服务器不存在")
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Host        *string `json:"host"`
		Port        *int    `json:"port"`
		Username    *string `json:"username"`
		AuthType    *string `json:"auth_type"`
		Password    *string `json:"password"`
		PrivateKey  *string `json:"private_key"`
		Description *string `json:"description"`
		Tags        *string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Name != nil {
		server.Name = *req.Name
	}
	if req.Host != nil {
		server.Host = *req.Host
	}
	if req.Port != nil {
		server.Port = *req.Port
	}
	if req.Username != nil {
		server.Username = *req.Username
	}
	if req.AuthType != nil {
		server.AuthType = *req.AuthType
	}
	if req.Password != nil {
		server.Password = *req.Password
	}
	if req.PrivateKey != nil {
		server.PrivateKey = *req.PrivateKey
	}
	if req.Description != nil {
		server.Description = *req.Description
	}
	if req.Tags != nil {
		server.Tags = *req.Tags
	}
	if err := h.serverService.Update(server); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	pkg.Success(c, server)
}

// DELETE /api/v1/servers/:id
func (h *ServerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.serverService.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

// POST /api/v1/servers/:id/test - test SSH connection
func (h *ServerHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	output, err := h.serverService.TestConnection(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, gin.H{"output": output})
}
