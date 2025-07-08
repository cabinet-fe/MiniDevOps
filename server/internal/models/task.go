package models

import (
	"time"

	"gorm.io/gorm"
)

// Task 任务模型
type Task struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name         string `json:"name" gorm:"not null" validate:"required"`             // 任务名称（必填）
	RepositoryID uint   `json:"repository_id" gorm:"not null" validate:"required"`    // 所属仓库（必填）
	Code         string `json:"code" gorm:"uniqueIndex;not null" validate:"required"` // 任务标识（必填，唯一）
	Branch       string `json:"branch" gorm:"not null" validate:"required"`           // 分支（必填）
	BuildScript  string `json:"build_script" gorm:"type:text"`                        // 构建脚本
	BuildPath    string `json:"build_path" gorm:""`                                   // 构建物路径
	AutoPush     bool   `json:"auto_push" gorm:"default:false"`                       // 构建后是否自动推送

	// 关联关系
	Repository    Repository     `json:"repository" gorm:"foreignKey:RepositoryID"`     // 所属仓库
	RemoteServers []RemoteServer `json:"remote_servers" gorm:"many2many:task_remotes;"` // 远程服务器列表
}

// TaskRemote 任务远程服务器关联表
type TaskRemote struct {
	TaskID         uint `json:"task_id" gorm:"primaryKey"`
	RemoteServerID uint `json:"remote_server_id" gorm:"primaryKey"`
}
