package handler

import (
	"io"
	"net/http"
	"strconv"

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
	rg.POST("/webhook/jobs/:build_job_id/:secret", h.Receive)
	rg.POST("/webhook/repos/:repository_id/:secret", h.ReceiveDeprecated)
}

func (h *WebhookHandler) Receive(c *gin.Context) {
	jobID, err := strconv.ParseUint(c.Param("build_job_id"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效构建任务 ID")
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

	result, err := h.svc.Receive(uint(jobID), secret, headers, body)
	if err != nil {
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

func (h *WebhookHandler) ReceiveDeprecated(c *gin.Context) {
	pkg.Error(c, http.StatusGone,
		"仓库级 Webhook 已弃用，请改用构建任务 Webhook：POST /api/v1/webhook/jobs/:build_job_id/:secret")
}
