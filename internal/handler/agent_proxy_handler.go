package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bedrock/internal/pkg"
	"bedrock/internal/service"
)

type AgentProxyHandler struct {
	proxyService *service.AgentProxyService
}

func NewAgentProxyHandler(ps *service.AgentProxyService) *AgentProxyHandler {
	return &AgentProxyHandler{proxyService: ps}
}

// GET /api/v1/agent-proxies
func (h *AgentProxyHandler) List(c *gin.Context) {
	pkg.Success(c, h.proxyService.List())
}

// POST /api/v1/agent-proxies/:key/install
func (h *AgentProxyHandler) Install(c *gin.Context) {
	key := c.Param("key")
	info, output, err := h.proxyService.Install(key)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, gin.H{"proxy": info, "output": output})
}

// POST /api/v1/agent-proxies/:key/upgrade
func (h *AgentProxyHandler) Upgrade(c *gin.Context) {
	key := c.Param("key")
	info, output, err := h.proxyService.Upgrade(key)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, gin.H{"proxy": info, "output": output})
}
