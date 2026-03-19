package middleware

import (
	"buildflow/internal/model"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Audit returns a middleware that records state-changing requests (POST, PUT, DELETE)
// to the audit_logs table. Captures user_id, action (derived from method + path), ip_address.
func Audit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
			c.Next()
			return
		}

		// Run the request first
		c.Next()

		// Record after response (user_id may be set by auth middleware)
		userID := GetUserID(c)
		action := method + " " + c.Request.URL.Path
		ip := c.ClientIP()

		entry := model.AuditLog{
			UserID:    userID,
			Action:    action,
			IPAddress: ip,
			CreatedAt: time.Now(),
		}

		_ = db.Create(&entry)
	}
}
