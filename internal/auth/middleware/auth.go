package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"bedrock/internal/auth/service"
	"bedrock/internal/pkg"
)

const (
	ctxUserID       = "user_id"
	ctxUsername     = "username"
	ctxIsSuperAdmin = "is_super_admin"
)

// Auth extracts Bearer JWT and sets user context.
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

		claims, err := authSvc.ParseToken(parts[1])
		if err != nil {
			pkg.Error(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Disabled users must fail subsequent requests.
		user, err := authSvc.GetByID(claims.UserID)
		if err != nil || user == nil || !user.IsActive {
			pkg.Error(c, http.StatusUnauthorized, "用户不存在或已被禁用")
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxIsSuperAdmin, user.IsSuperAdmin)
		c.Next()
	}
}

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

func IsSuperAdmin(c *gin.Context) bool {
	v, ok := c.Get(ctxIsSuperAdmin)
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}
