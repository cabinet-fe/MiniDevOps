package service

import (
	"strconv"

	"minidevops/server/internal/models"
	"minidevops/server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Username string `json:"username" validate:"required"` // 用户名（必填）
	Password string `json:"password" validate:"required"` // 密码（必填）
	Name     string `json:"name" validate:"required"`     // 名称（必填）
	Phone    string `json:"phone"`                        // 手机
	Email    string `json:"email"`                        // 邮箱
	RoleIDs  []uint `json:"role_ids"`                     // 角色ID列表
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Name     *string `json:"name"`     // 名称
	Phone    *string `json:"phone"`    // 手机
	Email    *string `json:"email"`    // 邮箱
	Password *string `json:"password"` // 密码
	RoleIDs  []uint  `json:"role_ids"` // 角色ID列表
}

// UserListResponse 用户列表响应结构
type UserListResponse struct {
	Total int64         `json:"total"`
	Items []models.User `json:"items"`
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))
	keyword := c.Query("keyword", "")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	query := s.db.Model(&models.User{}).Preload("Roles")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("username LIKE ? OR name LIKE ? OR phone LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询用户总数失败", err)
	}

	// 获取用户列表
	var users []models.User
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询用户列表失败", err)
	}

	response := UserListResponse{
		Total: total,
		Items: users,
	}

	return utils.SuccessWithData(c, "获取用户列表成功", response)
}

// CreateUser 创建用户
func (s *UserService) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查用户名是否已存在
	var existUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existUser).Error; err == nil {
		return utils.Error(c, fiber.StatusBadRequest, "用户名已存在", nil)
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "密码加密失败", err)
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Name:     req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "创建用户失败", err)
	}

	// 分配角色
	if len(req.RoleIDs) > 0 {
		if err := s.assignRoles(user.ID, req.RoleIDs); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "分配角色失败", err)
		}
	}

	// 重新加载用户信息（包含角色）
	if err := s.db.Preload("Roles").First(&user, user.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载用户信息失败", err)
	}

	return utils.SuccessWithData(c, "创建用户成功", user)
}

// GetUser 获取用户详情
func (s *UserService) GetUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "用户ID格式错误", err)
	}

	var user models.User
	if err := s.db.Preload("Roles.Permissions").First(&user, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "用户不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询用户失败", err)
	}

	return utils.SuccessWithData(c, "获取用户成功", user)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "用户ID格式错误", err)
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "用户不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询用户失败", err)
	}

	// 更新用户信息
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "密码加密失败", err)
		}
		updates["password"] = hashedPassword
	}

	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新用户失败", err)
		}
	}

	// 更新角色关联
	if req.RoleIDs != nil {
		if err := s.assignRoles(user.ID, req.RoleIDs); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新角色失败", err)
		}
	}

	// 重新加载用户信息
	if err := s.db.Preload("Roles").First(&user, user.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载用户信息失败", err)
	}

	return utils.SuccessWithData(c, "更新用户成功", user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "用户ID格式错误", err)
	}

	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "用户不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询用户失败", err)
	}

	// 删除用户
	if err := s.db.Delete(&user).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "删除用户失败", err)
	}

	return utils.Success(c, "删除用户成功")
}

// assignRoles 分配角色
func (s *UserService) assignRoles(userID uint, roleIDs []uint) error {
	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有的角色关联
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}

		// 创建新的角色关联
		for _, roleID := range roleIDs {
			userRole := models.UserRole{
				UserID: userID,
				RoleID: roleID,
			}
			if err := tx.Create(&userRole).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
