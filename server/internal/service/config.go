package service

import (
	"os"
	"path/filepath"

	"minidevops/server/internal/models"
	"minidevops/server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ConfigService 系统配置服务
type ConfigService struct {
	db *gorm.DB
}

// NewConfigService 创建系统配置服务实例
func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{db: db}
}

// ConfigRequest 配置请求结构
type ConfigRequest struct {
	Key         string `json:"key" validate:"required"` // 配置键
	Value       string `json:"value"`                   // 配置值
	Description string `json:"description"`             // 配置描述
}

// ConfigListResponse 配置列表响应结构
type ConfigListResponse struct {
	Total int64                 `json:"total"`
	Items []models.SystemConfig `json:"items"`
}

// GetConfigs 获取系统配置列表
func (s *ConfigService) GetConfigs(c *fiber.Ctx) error {
	var configs []models.SystemConfig
	if err := s.db.Order("created_at ASC").Find(&configs).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询配置列表失败", err)
	}

	response := ConfigListResponse{
		Total: int64(len(configs)),
		Items: configs,
	}

	return utils.SuccessWithData(c, "获取配置列表成功", response)
}

// GetConfig 获取指定配置
func (s *ConfigService) GetConfig(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return utils.Error(c, fiber.StatusBadRequest, "配置键不能为空", nil)
	}

	var config models.SystemConfig
	if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "配置不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询配置失败", err)
	}

	return utils.SuccessWithData(c, "获取配置成功", config)
}

// UpdateConfig 更新配置
func (s *ConfigService) UpdateConfig(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return utils.Error(c, fiber.StatusBadRequest, "配置键不能为空", nil)
	}

	var req ConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 查找现有配置
	var config models.SystemConfig
	err := s.db.Where("key = ?", key).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 配置不存在，创建新配置
		config = models.SystemConfig{
			Key:         key,
			Value:       req.Value,
			Description: req.Description,
		}
		if err := s.db.Create(&config).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "创建配置失败", err)
		}
	} else if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询配置失败", err)
	} else {
		// 配置存在，更新配置
		updates := map[string]interface{}{
			"value": req.Value,
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}

		if err := s.db.Model(&config).Updates(updates).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新配置失败", err)
		}

		// 重新加载配置
		if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "加载配置失败", err)
		}
	}

	return utils.SuccessWithData(c, "更新配置成功", config)
}

// DeleteConfig 删除配置
func (s *ConfigService) DeleteConfig(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return utils.Error(c, fiber.StatusBadRequest, "配置键不能为空", nil)
	}

	// 检查配置是否存在
	var config models.SystemConfig
	if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "配置不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询配置失败", err)
	}

	// 删除配置
	if err := s.db.Delete(&config).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "删除配置失败", err)
	}

	return utils.Success(c, "删除配置成功")
}

// GetMountPath 获取挂载路径配置
func (s *ConfigService) GetMountPath(c *fiber.Ctx) error {
	var config models.SystemConfig
	err := s.db.Where("key = ?", models.ConfigKeyMountPath).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 如果配置不存在，返回默认路径
		homeDir, _ := os.UserHomeDir()
		defaultPath := filepath.Join(homeDir, "dev-ops")

		response := map[string]interface{}{
			"key":         models.ConfigKeyMountPath,
			"value":       defaultPath,
			"description": "任务挂载路径",
			"is_default":  true,
		}
		return utils.SuccessWithData(c, "获取挂载路径成功", response)
	} else if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询挂载路径失败", err)
	}

	response := map[string]interface{}{
		"key":         config.Key,
		"value":       config.Value,
		"description": config.Description,
		"is_default":  false,
	}

	return utils.SuccessWithData(c, "获取挂载路径成功", response)
}

// UpdateMountPath 更新挂载路径配置
func (s *ConfigService) UpdateMountPath(c *fiber.Ctx) error {
	var req struct {
		Path string `json:"path" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 验证路径格式
	if !filepath.IsAbs(req.Path) {
		return utils.Error(c, fiber.StatusBadRequest, "路径必须是绝对路径", nil)
	}

	// 检查路径是否存在，如果不存在则尝试创建
	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		if err := os.MkdirAll(req.Path, 0755); err != nil {
			return utils.Error(c, fiber.StatusBadRequest, "无法创建指定路径", err)
		}
	}

	// 查找现有配置
	var config models.SystemConfig
	err := s.db.Where("key = ?", models.ConfigKeyMountPath).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 配置不存在，创建新配置
		config = models.SystemConfig{
			Key:         models.ConfigKeyMountPath,
			Value:       req.Path,
			Description: "任务挂载路径",
		}
		if err := s.db.Create(&config).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "创建挂载路径配置失败", err)
		}
	} else if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询挂载路径配置失败", err)
	} else {
		// 配置存在，更新配置
		if err := s.db.Model(&config).Update("value", req.Path).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新挂载路径配置失败", err)
		}

		// 重新加载配置
		if err := s.db.Where("key = ?", models.ConfigKeyMountPath).First(&config).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "加载挂载路径配置失败", err)
		}
	}

	return utils.SuccessWithData(c, "更新挂载路径成功", config)
}

// 初始化默认配置
func (s *ConfigService) InitDefaultConfigs() error {
	// 检查挂载路径配置是否存在
	var config models.SystemConfig
	err := s.db.Where("key = ?", models.ConfigKeyMountPath).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 创建默认挂载路径配置
		homeDir, _ := os.UserHomeDir()
		defaultPath := filepath.Join(homeDir, "dev-ops")

		config = models.SystemConfig{
			Key:         models.ConfigKeyMountPath,
			Value:       defaultPath,
			Description: "任务挂载路径",
		}

		if err := s.db.Create(&config).Error; err != nil {
			return err
		}

		// 创建目录
		os.MkdirAll(defaultPath, 0755)
	}

	return nil
}
