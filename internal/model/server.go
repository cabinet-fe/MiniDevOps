package model

import "time"

type Server struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Host        string    `json:"host" gorm:"size:200;not null"`
	Port        int       `json:"port" gorm:"default:22"`
	OSType      string    `json:"os_type" gorm:"size:20;not null;default:linux"`
	Username    string    `json:"username" gorm:"size:100;not null"`
	AuthType    string    `json:"auth_type" gorm:"size:20;not null"`
	Password    string    `json:"-" gorm:"size:1000"`
	PrivateKey  string    `json:"-" gorm:"type:text"`
	AgentURL    string    `json:"agent_url" gorm:"size:500"`
	AgentToken  string    `json:"-" gorm:"size:1000"`
	Description string    `json:"description" gorm:"size:500"`
	Tags        string    `json:"tags" gorm:"size:500"`
	Status      string    `json:"status" gorm:"size:20;default:unknown"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
