package service

import (
	"strconv"

	"server/internal/models"
	"server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RepositoryService 代码仓库服务
type RepositoryService struct {
	*CrudService[models.Repository]
}

// NewRepositoryService 创建代码仓库服务实例
func NewRepositoryService(db *gorm.DB) *RepositoryService {
	return &RepositoryService{CrudService: NewCrudService[models.Repository](db)}
}

// CreateRepositoryRequest 创建仓库请求结构
type CreateRepositoryRequest struct {
	Name   string `json:"name" validate:"required"` // 仓库名称（必填）
	URL    string `json:"url" validate:"required"`  // 仓库地址（必填）
	Branch string `json:"branch"`                   // 仓库分支
}

// UpdateRepositoryRequest 更新仓库请求结构
type UpdateRepositoryRequest struct {
	Name   *string `json:"name"`   // 仓库名称
	URL    *string `json:"url"`    // 仓库地址
	Branch *string `json:"branch"` // 仓库分支
}

// RepositoryListResponse 仓库列表响应结构
type RepositoryListResponse struct {
	Total int64               `json:"total"`
	Items []models.Repository `json:"items"`
}

// GetRepositories 获取仓库列表
func (s *RepositoryService) GetRepositories(c *fiber.Ctx) error {
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

	query := s.DB.Model(&models.Repository{}).Preload("Tasks")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR url LIKE ? OR branch LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询仓库总数失败", err)
	}

	// 获取仓库列表
	var repositories []models.Repository
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&repositories).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询仓库列表失败", err)
	}

	response := RepositoryListResponse{
		Total: total,
		Items: repositories,
	}

	return utils.SuccessWithData(c, "获取仓库列表成功", response)
}

// CreateRepository 创建仓库
func (s *RepositoryService) CreateRepository(c *fiber.Ctx) error {
	var req CreateRepositoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查仓库地址是否已存在
	var existRepository models.Repository
	if err := s.DB.Where("url = ?", req.URL).First(&existRepository).Error; err == nil {
		return utils.Error(c, fiber.StatusBadRequest, "仓库地址已存在", nil)
	}

	// 设置默认分支
	if req.Branch == "" {
		req.Branch = "main"
	}

	// 创建仓库
	repository := models.Repository{
		Name:   req.Name,
		URL:    req.URL,
		Branch: req.Branch,
	}

	if err := s.Create(c, &repository); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "创建仓库失败", err)
	}

	return utils.SuccessWithData(c, "创建仓库成功", repository)
}

// GetRepository 获取仓库详情
func (s *RepositoryService) GetRepository(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "仓库ID格式错误", err)
	}

	var repository models.Repository
	if err := s.DB.Preload("Tasks").First(&repository, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "仓库不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询仓库失败", err)
	}

	return utils.SuccessWithData(c, "获取仓库成功", repository)
}

// UpdateRepository 更新仓库
func (s *RepositoryService) UpdateRepository(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "仓库ID格式错误", err)
	}

	var req UpdateRepositoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查仓库是否存在
	repository, err := s.GetByID(c, uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "仓库不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询仓库失败", err)
	}

	// 检查仓库地址是否已存在（排除自己）
	if req.URL != nil {
		var existRepository models.Repository
		if err := s.DB.Where("url = ? AND id != ?", *req.URL, id).First(&existRepository).Error; err == nil {
			return utils.Error(c, fiber.StatusBadRequest, "仓库地址已存在", nil)
		}
	}

	// 更新仓库信息
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Branch != nil {
		updates["branch"] = *req.Branch
	}

	if len(updates) > 0 {
		if err := s.DB.Model(&repository).Updates(updates).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新仓库失败", err)
		}
	}

	// 重新加载仓库信息
	if err := s.DB.Preload("Tasks").First(&repository, repository.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载仓库信息失败", err)
	}

	return utils.SuccessWithData(c, "更新仓库成功", repository)
}

// DeleteRepository 删除仓库
func (s *RepositoryService) DeleteRepository(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "仓库ID格式错误", err)
	}

	// 检查仓库是否存在
	repository, err := s.GetByID(c, uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "仓库不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询仓库失败", err)
	}

	// 检查是否有任务使用该仓库
	if len(repository.Tasks) > 0 {
		return utils.Error(c, fiber.StatusBadRequest, "该仓库正在被任务使用，无法删除", nil)
	}

	// 删除仓库
	if err := s.Delete(c, uint(id)); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "删除仓库失败", err)
	}

	return utils.Success(c, "删除仓库成功")
}
