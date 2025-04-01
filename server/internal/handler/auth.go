package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"minidevops/internal/middleware"
	"minidevops/internal/model"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	client *model.Client
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(client *model.Client) *AuthHandler {
	return &AuthHandler{
		client: client,
	}
}

// UserCredentials 用户凭证
type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	GiteeToken string `json:"gitee_token"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token   string       `json:"token"`
	User    *UserProfile `json:"user"`
}

// UserProfile 用户资料
type UserProfile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Register godoc
// @Summary 注册新用户
// @Description 注册新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "用户注册信息"
// @Success 201 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的请求格式",
		})
	}

	// 验证请求数据
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "用户名和密码不能为空",
		})
	}

	// 检查用户名是否已存在
	exists, err := h.client.User.Query().Where(model.UserUsernameEQ(req.Username)).Exist(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "服务器内部错误",
		})
	}

	if exists {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "用户名已存在",
		})
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "密码处理错误",
		})
	}

	// 创建用户
	now := time.Now()
	user, err := h.client.User.
		Create().
		SetUsername(req.Username).
		SetPassword(string(hashedPassword)).
		SetGiteeToken(req.GiteeToken).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "创建用户失败",
		})
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "生成认证令牌失败",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(LoginResponse{
		Token: token,
		User: &UserProfile{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录并获取JWT令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param credentials body UserCredentials true "用户凭证"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var credentials UserCredentials
	if err := c.BodyParser(&credentials); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的请求格式",
		})
	}

	// 查找用户
	user, err := h.client.User.
		Query().
		Where(model.UserUsernameEQ(credentials.Username)).
		Only(c.Context())

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error: "用户名或密码错误",
		})
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error: "用户名或密码错误",
		})
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "生成认证令牌失败",
		})
	}

	return c.JSON(LoginResponse{
		Token: token,
		User: &UserProfile{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		},
	})
}