package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"bedrock/internal/cicd/service"
	"bedrock/internal/pkg"
)

type WebhookHandler struct {
	svc *service.WebhookService
}

func NewWebhookHandler(svc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

func (h *WebhookHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/webhook/repos/:repository_id/:secret", h.Receive)
}

func (h *WebhookHandler) Receive(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repository_id"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效仓库 ID")
		return
	}
	secret := c.Param("secret")
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 2<<20))
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无法读取请求体")
		return
	}

	headers := map[string]string{}
	for k, vals := range c.Request.Header {
		if len(vals) > 0 {
			headers[k] = vals[0]
		}
	}

	var filterJobID uint
	if q := strings.TrimSpace(c.Query("build_job_id")); q != "" {
		if id, err := strconv.ParseUint(q, 10, 64); err == nil {
			filterJobID = uint(id)
		}
	}

	result, err := h.svc.Receive(uint(repoID), secret, headers, body, filterJobID)
	if err != nil {
		// Never echo secret in error messages
		msg := service.RedactSecret(err.Error(), secret)
		switch {
		case service.IsUnauthorized(err):
			pkg.Error(c, http.StatusUnauthorized, msg)
		case service.IsNotFound(err):
			pkg.Error(c, http.StatusNotFound, msg)
		default:
			pkg.Error(c, http.StatusBadRequest, msg)
		}
		return
	}
	c.JSON(http.StatusAccepted, pkg.Response{Code: 0, Message: "accepted", Data: result})
}
