package models

import (
	"time"

	"gorm.io/gorm"
)

// AuthType 验证方式
type AuthType string

const (
	AuthTypePassword AuthType = "password" // 密码
	AuthTypeSSHKey   AuthType = "ssh-key"  // SSH密钥
)

// RemoteServer 远程服务器模型
type RemoteServer struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name     string   `json:"name" gorm:"not null" validate:"required"`      // 服务器名称（必填）
	Host     string   `json:"host" gorm:"not null" validate:"required"`      // 主机地址（必填）
	Port     int      `json:"port" gorm:"default:22"`                        // 端口
	AuthType AuthType `json:"auth_type" gorm:"not null" validate:"required"` // 验证方式
	Username string   `json:"username" gorm:"not null" validate:"required"`  // 用户名（必填）
	Password string   `json:"password,omitempty" gorm:""`                    // 密码（当验证方式为密码时）
	SSHKey   string   `json:"ssh_key,omitempty" gorm:"type:text"`            // SSH密钥（当验证方式为SSH密钥时）
	Path     string   `json:"path" gorm:"not null" validate:"required"`      // 路径（必填）

	// 关联关系
	Tasks []Task `json:"tasks" gorm:"many2many:task_remotes;"` // 关联的任务
}
