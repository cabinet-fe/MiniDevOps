package handler

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"buildflow/internal/config"
	"buildflow/internal/middleware"
	"buildflow/internal/repository"
	"buildflow/internal/service"
	"buildflow/internal/ws"
)

type WSHandler struct {
	authSvc     *service.AuthService
	buildRepo   *repository.BuildRepository
	projectRepo *repository.ProjectRepository
	hub         *ws.Hub
	cors        middleware.CORSConfig
}

func NewWSHandler(authSvc *service.AuthService, buildRepo *repository.BuildRepository, projectRepo *repository.ProjectRepository, hub *ws.Hub, cors middleware.CORSConfig) *WSHandler {
	return &WSHandler{
		authSvc:     authSvc,
		buildRepo:   buildRepo,
		projectRepo: projectRepo,
		hub:         hub,
		cors:        cors,
	}
}

func (h *WSHandler) wsUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return middleware.WebSocketCheckOrigin(h.cors, r)
		},
	}
}

// HandleBuildLogs upgrades to WebSocket, authenticates via query param "token",
// subscribes to channel "build:{id}", sends existing log file content first, then streams new lines.
func (h *WSHandler) HandleBuildLogs(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	claims, err := h.authSvc.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	buildIDStr := c.Param("id")
	buildID, err := strconv.ParseUint(buildIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid build id"})
		return
	}

	build, err := h.buildRepo.FindByID(uint(buildID))
	if err != nil || build == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "build not found"})
		return
	}

	// Check project access - user must have triggered or have access
	project, err := h.projectRepo.FindByID(build.ProjectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if !middleware.UserCanAccessProjectByIDs(claims.UserID, claims.Role, project.CreatedBy) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	conn, err := h.wsUpgrader().Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	channel := fmt.Sprintf("build:%d", build.ID)
	client := &ws.Client{
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Channel: channel,
		UserID:  claims.UserID,
	}
	h.hub.Register(client)
	go ws.WritePump(client, h.hub)

	// Send existing log file content first
	logPath := build.LogPath
	if logPath == "" && config.C != nil && config.C.Build.LogDir != "" {
		logPath = filepath.Join(config.C.Build.LogDir, fmt.Sprintf("project-%d", build.ProjectID), fmt.Sprintf("build-%03d.log", build.BuildNumber))
	} else if logPath != "" && !filepath.IsAbs(logPath) && config.C != nil && config.C.Build.LogDir != "" {
		logPath = filepath.Join(config.C.Build.LogDir, logPath)
	}
	if logPath != "" {
		if f, err := os.Open(logPath); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				select {
				case client.Send <- []byte(line):
				default:
					// Buffer full, skip
				}
			}
			f.Close()
		}
	}

	// ReadPump - discard incoming messages but keep connection alive
	go func() {
		defer h.hub.Unregister(client)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// HandleNotifications upgrades to WebSocket, authenticates via query param "token",
// subscribes to channel "notifications:{userID}".
func (h *WSHandler) HandleNotifications(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	claims, err := h.authSvc.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	conn, err := h.wsUpgrader().Upgrade(c.Writer, c.Request, nil)
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

	// ReadPump - discard incoming messages but keep connection alive
	go func() {
		defer h.hub.Unregister(client)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
