package model

import "time"

// User is the identity model. Roles are bound via user_roles (see internal/rbac).
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;size:255;not null"`
	DisplayName  string    `json:"display_name" gorm:"size:100"`
	Email        string    `json:"email" gorm:"size:200"`
	Avatar       string    `json:"avatar" gorm:"size:500"`
	IsActive     bool      `json:"is_active" gorm:"not null;default:true"`
	IsSuperAdmin bool      `json:"is_super_admin" gorm:"not null;default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }
