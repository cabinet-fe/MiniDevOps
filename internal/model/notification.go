package model

import "time"

type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Type      string    `json:"type" gorm:"size:30;not null"`
	Title     string    `json:"title" gorm:"size:200;not null"`
	Message   string    `json:"message" gorm:"size:500"`
	BuildID   *uint     `json:"build_id"`
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}
