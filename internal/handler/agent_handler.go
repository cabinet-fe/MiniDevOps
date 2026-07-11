package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

type AgentHandler struct {
	agentService *service.AgentService
}

func NewAgentHandler(as *service.AgentService) *AgentHandler {
	return &AgentHandler{agentService: as}
}

// GET /api/v1/agents
func (h *AgentHandler) List(c *gin.Context) {
	agents, err := h.agentService.List()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, agents)
}

// POST /api/v1/agents
func (h *AgentHandler) Create(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		Prompt     string `json:"prompt"`
		ProxyKey   string `json:"proxy_key" binding:"required"`
		Enabled    *bool  `json:"enabled"`
		ProjectIDs []uint `json:"project_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	agent := &model.Agent{
		Name:     req.Name,
		Prompt:   req.Prompt,
		ProxyKey: req.ProxyKey,
		Enabled:  enabled,
	}
	if err := h.agentService.Create(agent, req.ProjectIDs); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, agent)
}

// GET /api/v1/agents/:id
func (h *AgentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	agent, err := h.agentService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "智能体不存在")
		return
	}
	pkg.Success(c, agent)
}

// PUT /api/v1/agents/:id
func (h *AgentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	agent, err := h.agentService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "智能体不存在")
		return
	}
	var req struct {
		Name       *string `json:"name"`
		Prompt     *string `json:"prompt"`
		ProxyKey   *string `json:"proxy_key"`
		Enabled    *bool   `json:"enabled"`
		ProjectIDs []uint  `json:"project_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Prompt != nil {
		agent.Prompt = *req.Prompt
	}
	if req.ProxyKey != nil {
		agent.ProxyKey = *req.ProxyKey
	}
	if req.Enabled != nil {
		agent.Enabled = *req.Enabled
	}
	syncProjects := req.ProjectIDs != nil
	if err := h.agentService.Update(agent, req.ProjectIDs, syncProjects); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, agent)
}

// DELETE /api/v1/agents/:id
func (h *AgentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.agentService.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}
