package models

import (
	"gorm.io/gorm"
)

// Repository 代码仓库模型
type Repository struct {
	gorm.Model

	Name   string `json:"name" gorm:"not null" validate:"required"`      // 仓库名称（必填）
	URL    string `json:"url" gorm:"not null;index" validate:"required"` // 仓库地址（必填）
	Branch string `json:"branch" gorm:"default:'main'"`                  // 仓库分支

	// 关联关系
	Tasks []Task `json:"tasks" gorm:"foreignKey:RepositoryID"` // 关联的任务
}
