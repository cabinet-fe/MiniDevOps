package middleware

import (
	"net/http"

	"buildflow/internal/pkg"

	"github.com/gin-gonic/gin"
)

const ctxOwnerID = "owner_id"

// RequireRole returns a handler that checks if the user's role is in the allowed list.
// Returns 403 Forbidden if not authorized. Must be used after Auth middleware.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role := GetRole(c)
		if role == "" {
			pkg.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		if !allowed[role] {
			pkg.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		c.Next()
	}
}

// RequireOwnerOrRole returns a handler that checks if the user is the resource owner
// OR has one of the roles. Owner info is obtained from context key "owner_id" or "created_by".
// Returns 403 Forbidden if not authorized. Must be used after Auth middleware.
func RequireOwnerOrRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			pkg.Error(c, http.StatusForbidden, "forbidden")
			return
		}

		// Check if user has one of the allowed roles
		role := GetRole(c)
		if allowed[role] {
			c.Next()
			return
		}

		// Check if user is the resource owner
		var ownerID uint
		if v, ok := c.Get(ctxOwnerID); ok {
			if id, ok := v.(uint); ok {
				ownerID = id
			}
		}
		if v, ok := c.Get("created_by"); ok {
			if id, ok := v.(uint); ok {
				ownerID = id
			}
		}

		if ownerID != 0 && userID == ownerID {
			c.Next()
			return
		}

		pkg.Error(c, http.StatusForbidden, "forbidden")
	}
}

// SetOwnerID sets the resource owner ID in context for RequireOwnerOrRole to use.
func SetOwnerID(c *gin.Context, ownerID uint) {
	c.Set(ctxOwnerID, ownerID)
}
