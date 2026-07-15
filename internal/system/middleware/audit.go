package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	authmiddleware "bedrock/internal/auth/middleware"
	"bedrock/internal/system/service"
)

// AuditWrite records mutating API calls after successful responses.
func AuditWrite(audit *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
			return
		}
		if c.Writer.Status() >= 400 {
			return
		}
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		// Skip auth noise.
		if strings.HasPrefix(path, "/api/v1/auth/") {
			return
		}
		_ = audit.Write(
			authmiddleware.GetUserID(c),
			authmiddleware.GetUsername(c),
			method,
			path,
			c.Param("id"),
			"",
			c.ClientIP(),
		)
	}
}
