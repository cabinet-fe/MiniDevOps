package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Username string `json:"username" gorm:"uniqueIndex;not null;index" validate:"required"` // 用户名（必填）
	Password string `json:"-" gorm:"not null" validate:"required"`                          // 密码（必填）
	Name     string `json:"name" gorm:"not null" validate:"required"`                       // 名称（必填）
	Phone    string `json:"phone" gorm:""`                                                  // 手机
	Email    string `json:"email" gorm:""`                                                  // 邮箱

	// 关联关系
	Roles []Role `json:"roles" gorm:"many2many:user_roles;"` // 角色（可以有多个角色）
}

// UserRole 用户角色关联表
type UserRole struct {
	UserID uint `json:"user_id" gorm:"primaryKey;index"`
	RoleID uint `json:"role_id" gorm:"primaryKey;index"`
}
