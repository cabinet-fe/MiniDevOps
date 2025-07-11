package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Key         string `json:"key" gorm:"uniqueIndex;not null;index" validate:"required"` // 配置键
	Value       string `json:"value" gorm:"type:text"`                                    // 配置值
	Description string `json:"description" gorm:""`                                       // 配置描述
}

// 预定义的系统配置键
const (
	ConfigKeyMountPath = "mount_path" // 挂载路径配置
)
