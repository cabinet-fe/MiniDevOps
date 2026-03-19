package middleware

import (
	"net/http"
	"strings"

	"buildflow/internal/pkg"
	"buildflow/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ctxUserID   = "user_id"
	ctxUsername = "username"
	ctxRole     = "role"
)

// Auth extracts Bearer token from Authorization header, validates JWT,
// and sets user info (id, username, role) in gin.Context.
// Returns 401 if invalid or missing token.
func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			pkg.Error(c, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			pkg.Error(c, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		tokenString := parts[1]
		claims, err := authSvc.ParseToken(tokenString)
		if err != nil {
			pkg.Error(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// GetUserID returns the authenticated user ID from context (0 if not set).
func GetUserID(c *gin.Context) uint {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return 0
	}
	if id, ok := v.(uint); ok {
		return id
	}
	return 0
}

// GetUsername returns the authenticated username from context.
func GetUsername(c *gin.Context) string {
	v, ok := c.Get(ctxUsername)
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// GetRole returns the authenticated user role from context.
func GetRole(c *gin.Context) string {
	v, ok := c.Get(ctxRole)
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
