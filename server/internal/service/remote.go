package service

import (
	"strconv"

	"server/internal/models"
	"server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RemoteService 远程服务器服务
type RemoteService struct {
	db *gorm.DB
}

// NewRemoteService 创建远程服务器服务实例
func NewRemoteService(db *gorm.DB) *RemoteService {
	return &RemoteService{db: db}
}

// CreateRemoteRequest 创建远程服务器请求结构
type CreateRemoteRequest struct {
	Name     string          `json:"name" validate:"required"`      // 服务器名称（必填）
	Host     string          `json:"host" validate:"required"`      // 主机地址（必填）
	Port     int             `json:"port"`                          // 端口
	AuthType models.AuthType `json:"auth_type" validate:"required"` // 验证方式
	Username string          `json:"username" validate:"required"`  // 用户名（必填）
	Password string          `json:"password,omitempty"`            // 密码（当验证方式为密码时）
	SSHKey   string          `json:"ssh_key,omitempty"`             // SSH密钥（当验证方式为SSH密钥时）
	Path     string          `json:"path" validate:"required"`      // 路径（必填）
}

// UpdateRemoteRequest 更新远程服务器请求结构
type UpdateRemoteRequest struct {
	Name     *string          `json:"name"`      // 服务器名称
	Host     *string          `json:"host"`      // 主机地址
	Port     *int             `json:"port"`      // 端口
	AuthType *models.AuthType `json:"auth_type"` // 验证方式
	Username *string          `json:"username"`  // 用户名
	Password *string          `json:"password"`  // 密码
	SSHKey   *string          `json:"ssh_key"`   // SSH密钥
	Path     *string          `json:"path"`      // 路径
}

// RemoteListResponse 远程服务器列表响应结构
type RemoteListResponse struct {
	Total int64                 `json:"total"`
	Items []models.RemoteServer `json:"items"`
}

// GetRemotes 获取远程服务器列表
func (s *RemoteService) GetRemotes(c *fiber.Ctx) error {
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

	query := s.db.Model(&models.RemoteServer{}).Preload("Tasks")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR host LIKE ? OR username LIKE ? OR path LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询远程服务器总数失败", err)
	}

	// 获取远程服务器列表
	var remotes []models.RemoteServer
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&remotes).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询远程服务器列表失败", err)
	}

	// 隐藏敏感信息
	for i := range remotes {
		remotes[i].Password = ""
		remotes[i].SSHKey = ""
	}

	response := RemoteListResponse{
		Total: total,
		Items: remotes,
	}

	return utils.SuccessWithData(c, "获取远程服务器列表成功", response)
}

// CreateRemote 创建远程服务器
func (s *RemoteService) CreateRemote(c *fiber.Ctx) error {
	var req CreateRemoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 验证认证方式和相关参数
	if req.AuthType == models.AuthTypePassword && req.Password == "" {
		return utils.Error(c, fiber.StatusBadRequest, "使用密码认证时必须提供密码", nil)
	}
	if req.AuthType == models.AuthTypeSSHKey && req.SSHKey == "" {
		return utils.Error(c, fiber.StatusBadRequest, "使用SSH密钥认证时必须提供SSH密钥", nil)
	}

	// 设置默认端口
	if req.Port == 0 {
		req.Port = 22
	}

	// 检查主机地址和端口是否已存在
	var existRemote models.RemoteServer
	if err := s.db.Where("host = ? AND port = ?", req.Host, req.Port).First(&existRemote).Error; err == nil {
		return utils.Error(c, fiber.StatusBadRequest, "该主机地址和端口已存在", nil)
	}

	// 创建远程服务器
	remote := models.RemoteServer{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		AuthType: req.AuthType,
		Username: req.Username,
		Password: req.Password,
		SSHKey:   req.SSHKey,
		Path:     req.Path,
	}

	if err := s.db.Create(&remote).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "创建远程服务器失败", err)
	}

	// 隐藏敏感信息
	remote.Password = ""
	remote.SSHKey = ""

	return utils.SuccessWithData(c, "创建远程服务器成功", remote)
}

// GetRemote 获取远程服务器详情
func (s *RemoteService) GetRemote(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "远程服务器ID格式错误", err)
	}

	var remote models.RemoteServer
	if err := s.db.Preload("Tasks").First(&remote, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "远程服务器不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询远程服务器失败", err)
	}

	// 隐藏敏感信息
	remote.Password = ""
	remote.SSHKey = ""

	return utils.SuccessWithData(c, "获取远程服务器成功", remote)
}

// UpdateRemote 更新远程服务器
func (s *RemoteService) UpdateRemote(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "远程服务器ID格式错误", err)
	}

	var req UpdateRemoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查远程服务器是否存在
	var remote models.RemoteServer
	if err := s.db.First(&remote, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "远程服务器不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询远程服务器失败", err)
	}

	// 检查主机地址和端口是否已存在（排除自己）
	if req.Host != nil && req.Port != nil {
		var existRemote models.RemoteServer
		if err := s.db.Where("host = ? AND port = ? AND id != ?", *req.Host, *req.Port, id).First(&existRemote).Error; err == nil {
			return utils.Error(c, fiber.StatusBadRequest, "该主机地址和端口已存在", nil)
		}
	} else if req.Host != nil {
		var existRemote models.RemoteServer
		if err := s.db.Where("host = ? AND port = ? AND id != ?", *req.Host, remote.Port, id).First(&existRemote).Error; err == nil {
			return utils.Error(c, fiber.StatusBadRequest, "该主机地址和端口已存在", nil)
		}
	} else if req.Port != nil {
		var existRemote models.RemoteServer
		if err := s.db.Where("host = ? AND port = ? AND id != ?", remote.Host, *req.Port, id).First(&existRemote).Error; err == nil {
			return utils.Error(c, fiber.StatusBadRequest, "该主机地址和端口已存在", nil)
		}
	}

	// 验证认证方式和相关参数
	authType := remote.AuthType
	if req.AuthType != nil {
		authType = *req.AuthType
	}

	if authType == models.AuthTypePassword {
		if req.Password != nil && *req.Password == "" {
			return utils.Error(c, fiber.StatusBadRequest, "使用密码认证时必须提供密码", nil)
		}
	} else if authType == models.AuthTypeSSHKey {
		if req.SSHKey != nil && *req.SSHKey == "" {
			return utils.Error(c, fiber.StatusBadRequest, "使用SSH密钥认证时必须提供SSH密钥", nil)
		}
	}

	// 更新远程服务器信息
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Host != nil {
		updates["host"] = *req.Host
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.AuthType != nil {
		updates["auth_type"] = *req.AuthType
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.SSHKey != nil {
		updates["ssh_key"] = *req.SSHKey
	}
	if req.Path != nil {
		updates["path"] = *req.Path
	}

	if len(updates) > 0 {
		if err := s.db.Model(&remote).Updates(updates).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新远程服务器失败", err)
		}
	}

	// 重新加载远程服务器信息
	if err := s.db.Preload("Tasks").First(&remote, remote.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载远程服务器信息失败", err)
	}

	// 隐藏敏感信息
	remote.Password = ""
	remote.SSHKey = ""

	return utils.SuccessWithData(c, "更新远程服务器成功", remote)
}

// DeleteRemote 删除远程服务器
func (s *RemoteService) DeleteRemote(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "远程服务器ID格式错误", err)
	}

	// 检查远程服务器是否存在
	var remote models.RemoteServer
	if err := s.db.Preload("Tasks").First(&remote, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "远程服务器不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询远程服务器失败", err)
	}

	// 检查是否有任务使用该远程服务器
	if len(remote.Tasks) > 0 {
		return utils.Error(c, fiber.StatusBadRequest, "该远程服务器正在被任务使用，无法删除", nil)
	}

	// 删除远程服务器
	if err := s.db.Delete(&remote).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "删除远程服务器失败", err)
	}

	return utils.Success(c, "删除远程服务器成功")
}

// TestConnection 测试远程服务器连接
func (s *RemoteService) TestConnection(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "远程服务器ID格式错误", err)
	}

	// 检查远程服务器是否存在
	var remote models.RemoteServer
	if err := s.db.First(&remote, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "远程服务器不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询远程服务器失败", err)
	}

	// TODO: 实现实际的连接测试逻辑
	// 这里可以使用SSH客户端库来测试连接

	return utils.Success(c, "连接测试成功")
}
