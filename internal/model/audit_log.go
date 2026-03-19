package model

import "time"

type AuditLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"index"`
	Action       string    `json:"action" gorm:"size:50;not null;index"`
	ResourceType string    `json:"resource_type" gorm:"size:30"`
	ResourceID   uint      `json:"resource_id"`
	Details      string    `json:"details" gorm:"type:text"`
	IPAddress    string    `json:"ip_address" gorm:"size:45"`
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
}
