package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	authservice "bedrock/internal/auth/service"
	"bedrock/internal/middleware"
	"bedrock/internal/ws"
)

// NotificationWSHandler streams per-user notifications (channel notifications:{userId}).
type NotificationWSHandler struct {
	auth *authservice.AuthService
	hub  *ws.Hub
	cors middleware.CORSConfig
}

func NewNotificationWSHandler(auth *authservice.AuthService, hub *ws.Hub, cors middleware.CORSConfig) *NotificationWSHandler {
	return &NotificationWSHandler{auth: auth, hub: hub, cors: cors}
}

func (h *NotificationWSHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/ws/notifications", h.HandleNotifications)
}

func (h *NotificationWSHandler) HandleNotifications(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	claims, err := h.auth.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return middleware.WebSocketCheckOrigin(h.cors, r)
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	channel := fmt.Sprintf("notifications:%d", claims.UserID)
	client := &ws.Client{
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Channel: channel,
		UserID:  claims.UserID,
	}
	h.hub.Register(client)
	go ws.WritePump(client, h.hub)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	h.hub.Unregister(client)
}
