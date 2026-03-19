package model

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	DisplayName  string    `json:"display_name" gorm:"size:100"`
	Role         string    `json:"role" gorm:"size:20;not null;default:dev"`
	Email        string    `json:"email" gorm:"size:200"`
	Avatar       string    `json:"avatar" gorm:"size:500"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
