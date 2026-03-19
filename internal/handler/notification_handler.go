package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"buildflow/internal/middleware"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(ns *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: ns}
}

// GET /api/v1/notifications - list by user
func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, pageSize := pkg.GetPage(c)
	notifications, total, err := h.notificationService.ListByUser(userID, page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Paginated(c, notifications, total, page, pageSize)
}

// PUT /api/v1/notifications/:id/read - mark read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.notificationService.MarkRead(uint(id), userID); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "操作失败")
		return
	}
	pkg.Success(c, nil)
}

// PUT /api/v1/notifications/read-all - mark all read
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.notificationService.MarkAllRead(userID); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "操作失败")
		return
	}
	pkg.Success(c, nil)
}
