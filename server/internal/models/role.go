package models

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string `json:"name" gorm:"not null" validate:"required"`                   // 角色名称（必填）
	Code        string `json:"code" gorm:"uniqueIndex;not null;index" validate:"required"` // 角色标识（必填）
	Description string `json:"description" gorm:"not null" validate:"required"`            // 角色描述（必填）
	DataScope   string `json:"data_scope" gorm:"default:'1'"`                              // 数据权限

	// 关联关系
	Users       []User       `json:"users" gorm:"many2many:user_roles;"`             // 用户列表
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"` // 权限列表
}

// RolePermission 角色权限关联表
type RolePermission struct {
	RoleID       uint `json:"role_id" gorm:"primaryKey;index"`
	PermissionID uint `json:"permission_id" gorm:"primaryKey;index"`
}
