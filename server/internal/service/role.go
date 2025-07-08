package service

import (
	"strconv"

	"minidevops/server/internal/models"
	"minidevops/server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RoleService 角色服务
type RoleService struct {
	db *gorm.DB
}

// NewRoleService 创建角色服务实例
func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// CreateRoleRequest 创建角色请求结构
type CreateRoleRequest struct {
	Name          string `json:"name" validate:"required"`        // 角色名称（必填）
	Code          string `json:"code" validate:"required"`        // 角色标识（必填）
	Description   string `json:"description" validate:"required"` // 角色描述（必填）
	DataScope     string `json:"data_scope"`                      // 数据权限
	PermissionIDs []uint `json:"permission_ids"`                  // 权限ID列表
}

// UpdateRoleRequest 更新角色请求结构
type UpdateRoleRequest struct {
	Name          *string `json:"name"`           // 角色名称
	Code          *string `json:"code"`           // 角色标识
	Description   *string `json:"description"`    // 角色描述
	DataScope     *string `json:"data_scope"`     // 数据权限
	PermissionIDs []uint  `json:"permission_ids"` // 权限ID列表
}

// RoleListResponse 角色列表响应结构
type RoleListResponse struct {
	Total int64         `json:"total"`
	Items []models.Role `json:"items"`
}

// GetRoles 获取角色列表
func (s *RoleService) GetRoles(c *fiber.Ctx) error {
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

	query := s.db.Model(&models.Role{}).Preload("Permissions")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询角色总数失败", err)
	}

	// 获取角色列表
	var roles []models.Role
	if err := query.Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询角色列表失败", err)
	}

	response := RoleListResponse{
		Total: total,
		Items: roles,
	}

	return utils.SuccessWithData(c, "获取角色列表成功", response)
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(c *fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查角色标识是否已存在
	var existRole models.Role
	if err := s.db.Where("code = ?", req.Code).First(&existRole).Error; err == nil {
		return utils.Error(c, fiber.StatusBadRequest, "角色标识已存在", nil)
	}

	// 创建角色
	role := models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		DataScope:   req.DataScope,
	}

	if err := s.db.Create(&role).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "创建角色失败", err)
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		if err := s.assignPermissions(role.ID, req.PermissionIDs); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "分配权限失败", err)
		}
	}

	// 重新加载角色信息（包含权限）
	if err := s.db.Preload("Permissions").First(&role, role.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载角色信息失败", err)
	}

	return utils.SuccessWithData(c, "创建角色成功", role)
}

// GetRole 获取角色详情
func (s *RoleService) GetRole(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "角色ID格式错误", err)
	}

	var role models.Role
	if err := s.db.Preload("Permissions").Preload("Users").First(&role, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "角色不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询角色失败", err)
	}

	return utils.SuccessWithData(c, "获取角色成功", role)
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "角色ID格式错误", err)
	}

	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查角色是否存在
	var role models.Role
	if err := s.db.First(&role, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "角色不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询角色失败", err)
	}

	// 检查角色标识是否已存在（排除自己）
	if req.Code != nil {
		var existRole models.Role
		if err := s.db.Where("code = ? AND id != ?", *req.Code, id).First(&existRole).Error; err == nil {
			return utils.Error(c, fiber.StatusBadRequest, "角色标识已存在", nil)
		}
	}

	// 更新角色信息
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.DataScope != nil {
		updates["data_scope"] = *req.DataScope
	}

	if len(updates) > 0 {
		if err := s.db.Model(&role).Updates(updates).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新角色失败", err)
		}
	}

	// 更新权限关联
	if req.PermissionIDs != nil {
		if err := s.assignPermissions(role.ID, req.PermissionIDs); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新权限失败", err)
		}
	}

	// 重新加载角色信息
	if err := s.db.Preload("Permissions").First(&role, role.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载角色信息失败", err)
	}

	return utils.SuccessWithData(c, "更新角色成功", role)
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "角色ID格式错误", err)
	}

	// 检查角色是否存在
	var role models.Role
	if err := s.db.Preload("Users").First(&role, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "角色不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询角色失败", err)
	}

	// 检查是否有用户使用该角色
	if len(role.Users) > 0 {
		return utils.Error(c, fiber.StatusBadRequest, "该角色正在被用户使用，无法删除", nil)
	}

	// 删除角色
	if err := s.db.Delete(&role).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "删除角色失败", err)
	}

	return utils.Success(c, "删除角色成功")
}

// assignPermissions 分配权限
func (s *RoleService) assignPermissions(roleID uint, permissionIDs []uint) error {
	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有的权限关联
		if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}

		// 创建新的权限关联
		for _, permissionID := range permissionIDs {
			rolePermission := models.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
			}
			if err := tx.Create(&rolePermission).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
