package middleware

import (
	"buildflow/internal/model"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var methodActionMap = map[string]string{
	"POST":   "create",
	"PUT":    "update",
	"PATCH":  "update",
	"DELETE": "delete",
}

var segmentResourceMap = map[string]string{
	"projects":     "project",
	"servers":      "server",
	"builds":       "build",
	"environments": "environment",
	"users":        "user",
	"settings":     "settings",
	"system":       "system",
}

func Audit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
			c.Next()
			return
		}

		c.Next()

		userID := GetUserID(c)
		ip := c.ClientIP()
		path := c.Request.URL.Path

		action, resourceType, resourceID := parseAuditInfo(method, path)

		entry := model.AuditLog{
			UserID:       userID,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Details:      method + " " + path,
			IPAddress:    ip,
			CreatedAt:    time.Now(),
		}

		_ = db.Create(&entry)
	}
}

func parseAuditInfo(method, path string) (action, resourceType string, resourceID uint) {
	if strings.Contains(path, "/auth/login") {
		return "login", "auth", 0
	}
	if strings.Contains(path, "/auth/") {
		return "auth", "auth", 0
	}

	action = methodActionMap[method]
	if action == "" {
		action = strings.ToLower(method)
	}

	trimmed := strings.TrimPrefix(path, "/api/v1/")
	segments := strings.Split(trimmed, "/")

	for _, seg := range segments {
		if rt, ok := segmentResourceMap[seg]; ok {
			resourceType = rt
			break
		}
	}
	if resourceType == "" && len(segments) > 0 {
		resourceType = strings.TrimRight(segments[0], "s")
	}

	for i := len(segments) - 1; i >= 0; i-- {
		if id, err := strconv.ParseUint(segments[i], 10, 32); err == nil && id > 0 {
			resourceID = uint(id)
			break
		}
	}

	if strings.Contains(path, "/test") {
		action = "test"
	}

	return
}
