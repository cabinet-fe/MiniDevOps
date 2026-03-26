package middleware

import (
	"github.com/gin-gonic/gin"
)

// UserCanAccessProject reports whether the user may access resources tied to a project
// (same rule as project list: admin/ops see all; dev only projects they created).
func UserCanAccessProject(c *gin.Context, projectCreatedBy uint) bool {
	return UserCanAccessProjectByIDs(GetUserID(c), GetRole(c), projectCreatedBy)
}

// UserCanAccessProjectByIDs is the same rule without gin.Context (e.g. WebSocket before upgrade).
func UserCanAccessProjectByIDs(userID uint, role string, projectCreatedBy uint) bool {
	switch role {
	case "admin", "ops":
		return true
	case "dev":
		return userID != 0 && userID == projectCreatedBy
	default:
		return false
	}
}
