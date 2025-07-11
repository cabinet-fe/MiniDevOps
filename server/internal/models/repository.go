package models

import (
	"time"

	"gorm.io/gorm"
)

// Repository 代码仓库模型
type Repository struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name   string `json:"name" gorm:"not null" validate:"required"`      // 仓库名称（必填）
	URL    string `json:"url" gorm:"not null;index" validate:"required"` // 仓库地址（必填）
	Branch string `json:"branch" gorm:"default:'main'"`                  // 仓库分支

	// 关联关系
	Tasks []Task `json:"tasks" gorm:"foreignKey:RepositoryID"` // 关联的任务
}
