package service

import (
	"strconv"

	"minidevops/server/internal/models"
	"minidevops/server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PermissionService 权限服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// CreatePermissionRequest 创建权限请求结构
type CreatePermissionRequest struct {
	Name     string                `json:"name" validate:"required"` // 权限名称（必填）
	Type     models.PermissionType `json:"type" validate:"required"` // 类型（菜单、按钮）
	Code     string                `json:"code" validate:"required"` // 权限标识（必填）
	Sort     int                   `json:"sort"`                     // 排序
	ParentID *uint                 `json:"parent_id"`                // 父级菜单ID
}

// UpdatePermissionRequest 更新权限请求结构
type UpdatePermissionRequest struct {
	Name     *string                `json:"name"`      // 权限名称
	Type     *models.PermissionType `json:"type"`      // 类型
	Code     *string                `json:"code"`      // 权限标识
	Sort     *int                   `json:"sort"`      // 排序
	ParentID *uint                  `json:"parent_id"` // 父级菜单ID
}

// PermissionListResponse 权限列表响应结构
type PermissionListResponse struct {
	Total int64               `json:"total"`
	Items []models.Permission `json:"items"`
}

// GetPermissions 获取权限列表
func (s *PermissionService) GetPermissions(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))
	keyword := c.Query("keyword", "")
	permType := c.Query("type", "")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	query := s.db.Model(&models.Permission{}).Preload("Parent").Preload("Children")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 类型筛选
	if permType != "" {
		query = query.Where("type = ?", permType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询权限总数失败", err)
	}

	// 获取权限列表
	var permissions []models.Permission
	if err := query.Order("sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&permissions).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询权限列表失败", err)
	}

	response := PermissionListResponse{
		Total: total,
		Items: permissions,
	}

	return utils.SuccessWithData(c, "获取权限列表成功", response)
}

// GetPermissionTree 获取权限树形结构
func (s *PermissionService) GetPermissionTree(c *fiber.Ctx) error {
	var permissions []models.Permission
	if err := s.db.Preload("Children").Where("parent_id IS NULL").Order("sort ASC, id ASC").Find(&permissions).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询权限树失败", err)
	}

	// 递归加载子权限
	for i := range permissions {
		s.loadChildrenRecursively(&permissions[i])
	}

	return utils.SuccessWithData(c, "获取权限树成功", permissions)
}

// CreatePermission 创建权限
func (s *PermissionService) CreatePermission(c *fiber.Ctx) error {
	var req CreatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查权限标识是否已存在
	var existPermission models.Permission
	if err := s.db.Where("code = ?", req.Code).First(&existPermission).Error; err == nil {
		return utils.Error(c, fiber.StatusBadRequest, "权限标识已存在", nil)
	}

	// 如果有父级权限，检查父级权限是否存在且为菜单类型
	if req.ParentID != nil {
		var parentPermission models.Permission
		if err := s.db.First(&parentPermission, *req.ParentID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return utils.Error(c, fiber.StatusBadRequest, "父级权限不存在", nil)
			}
			return utils.Error(c, fiber.StatusInternalServerError, "查询父级权限失败", err)
		}

		// 只有菜单类型的权限可以作为父级
		if parentPermission.Type != models.PermissionTypeMenu {
			return utils.Error(c, fiber.StatusBadRequest, "只有菜单类型的权限可以作为父级", nil)
		}
	}

	// 创建权限
	permission := models.Permission{
		Name:     req.Name,
		Type:     req.Type,
		Code:     req.Code,
		Sort:     req.Sort,
		ParentID: req.ParentID,
	}

	if err := s.db.Create(&permission).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "创建权限失败", err)
	}

	// 重新加载权限信息（包含父级和子级）
	if err := s.db.Preload("Parent").Preload("Children").First(&permission, permission.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载权限信息失败", err)
	}

	return utils.SuccessWithData(c, "创建权限成功", permission)
}

// GetPermission 获取权限详情
func (s *PermissionService) GetPermission(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "权限ID格式错误", err)
	}

	var permission models.Permission
	if err := s.db.Preload("Parent").Preload("Children").Preload("Roles").First(&permission, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "权限不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询权限失败", err)
	}

	return utils.SuccessWithData(c, "获取权限成功", permission)
}

// UpdatePermission 更新权限
func (s *PermissionService) UpdatePermission(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "权限ID格式错误", err)
	}

	var req UpdatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查权限是否存在
	var permission models.Permission
	if err := s.db.First(&permission, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "权限不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询权限失败", err)
	}

	// 检查权限标识是否已存在（排除自己）
	if req.Code != nil {
		var existPermission models.Permission
		if err := s.db.Where("code = ? AND id != ?", *req.Code, id).First(&existPermission).Error; err == nil {
			return utils.Error(c, fiber.StatusBadRequest, "权限标识已存在", nil)
		}
	}

	// 如果要更新父级权限，进行相关检查
	if req.ParentID != nil {
		// 不能将自己设置为父级
		if *req.ParentID == uint(id) {
			return utils.Error(c, fiber.StatusBadRequest, "不能将自己设置为父级权限", nil)
		}

		// 检查父级权限是否存在且为菜单类型
		if *req.ParentID != 0 {
			var parentPermission models.Permission
			if err := s.db.First(&parentPermission, *req.ParentID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return utils.Error(c, fiber.StatusBadRequest, "父级权限不存在", nil)
				}
				return utils.Error(c, fiber.StatusInternalServerError, "查询父级权限失败", err)
			}

			if parentPermission.Type != models.PermissionTypeMenu {
				return utils.Error(c, fiber.StatusBadRequest, "只有菜单类型的权限可以作为父级", nil)
			}
		} else {
			// 设置为根权限
			req.ParentID = nil
		}
	}

	// 更新权限信息
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}

	if len(updates) > 0 {
		if err := s.db.Model(&permission).Updates(updates).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新权限失败", err)
		}
	}

	// 重新加载权限信息
	if err := s.db.Preload("Parent").Preload("Children").First(&permission, permission.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载权限信息失败", err)
	}

	return utils.SuccessWithData(c, "更新权限成功", permission)
}

// DeletePermission 删除权限
func (s *PermissionService) DeletePermission(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "权限ID格式错误", err)
	}

	// 检查权限是否存在
	var permission models.Permission
	if err := s.db.Preload("Children").First(&permission, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "权限不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询权限失败", err)
	}

	// 检查是否有子权限
	if len(permission.Children) > 0 {
		return utils.Error(c, fiber.StatusBadRequest, "该权限下还有子权限，无法删除", nil)
	}

	// 删除权限
	if err := s.db.Delete(&permission).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "删除权限失败", err)
	}

	return utils.Success(c, "删除权限成功")
}

// loadChildrenRecursively 递归加载子权限
func (s *PermissionService) loadChildrenRecursively(permission *models.Permission) {
	var children []models.Permission
	if err := s.db.Where("parent_id = ?", permission.ID).Order("sort ASC, id ASC").Find(&children).Error; err != nil {
		return
	}

	permission.Children = children
	for i := range permission.Children {
		s.loadChildrenRecursively(&permission.Children[i])
	}
}
