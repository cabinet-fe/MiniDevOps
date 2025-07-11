package models

import (
	"time"

	"gorm.io/gorm"
)

// PermissionType 权限类型
type PermissionType string

const (
	PermissionTypeMenu   PermissionType = "menu"   // 菜单
	PermissionTypeButton PermissionType = "button" // 按钮
)

// Permission 权限模型
type Permission struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name     string         `json:"name" gorm:"not null" validate:"required"`                   // 权限名称（必填）
	Type     PermissionType `json:"type" gorm:"not null" validate:"required"`                   // 类型（菜单、按钮）
	Code     string         `json:"code" gorm:"uniqueIndex;not null;index" validate:"required"` // 权限标识（必填）
	Sort     int            `json:"sort" gorm:"default:0"`                                      // 排序
	ParentID *uint          `json:"parent_id" gorm:"default:null;index"`                        // 父级菜单ID

	// 关联关系
	Parent   *Permission  `json:"parent" gorm:"foreignKey:ParentID"`        // 父级权限
	Children []Permission `json:"children" gorm:"foreignKey:ParentID"`      // 子权限
	Roles    []Role       `json:"roles" gorm:"many2many:role_permissions;"` // 关联的角色
}
