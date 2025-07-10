package service

import (
	"fmt"
	"server/internal/models"
	"server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// Login 用户登录
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	fmt.Println(req)

	// 查找用户
	var user models.User
	if err := s.db.Preload("Roles.Permissions").Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusUnauthorized, "用户名或密码错误", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "数据库查询失败", err)
	}

	// 验证密码
	if err := utils.VerifyPassword(user.Password, req.Password); err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "用户名或密码错误", nil)
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "生成令牌失败", err)
	}

	response := LoginResponse{
		Token: token,
		User:  user,
	}

	return utils.SuccessWithData(c, "登录成功", response)
}

// Logout 用户登出
func (s *AuthService) Logout(c *fiber.Ctx) error {
	// 在实际应用中，可以将token加入黑名单
	return utils.Success(c, "登出成功")
}

// GetProfile 获取用户信息
func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	// 从JWT中获取用户ID
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		return utils.Error(c, fiber.StatusUnauthorized, "用户未认证", nil)
	}

	// 查找用户信息
	var user models.User
	if err := s.db.Preload("Roles.Permissions").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "用户不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "数据库查询失败", err)
	}

	return utils.SuccessWithData(c, "获取用户信息成功", user)
}
