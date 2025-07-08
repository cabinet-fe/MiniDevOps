package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware JWT中间件
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取Authorization头
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return Error(c, fiber.StatusUnauthorized, "缺少认证令牌", nil)
		}

		// 检查是否以"Bearer "开头
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return Error(c, fiber.StatusUnauthorized, "认证令牌格式错误", nil)
		}

		// 提取令牌
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return Error(c, fiber.StatusUnauthorized, "认证令牌为空", nil)
		}

		// 解析令牌
		claims, err := ParseToken(token)
		if err != nil {
			return Error(c, fiber.StatusUnauthorized, "认证令牌无效", err)
		}

		// 将用户信息存储到上下文中
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)

		return c.Next()
	}
}

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(c *fiber.Ctx) uint {
	if userID, ok := c.Locals("user_id").(uint); ok {
		return userID
	}
	return 0
}

// GetUsernameFromContext 从上下文中获取用户名
func GetUsernameFromContext(c *fiber.Ctx) string {
	if username, ok := c.Locals("username").(string); ok {
		return username
	}
	return ""
}
