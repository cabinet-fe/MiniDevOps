package model

import "time"

// Distribution is one deploy target for an environment (multi-distribution).
type Distribution struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	EnvironmentID    uint      `json:"environment_id" gorm:"not null;index"`
	ServerID         *uint     `json:"server_id"`
	RemotePath       string    `json:"remote_path" gorm:"size:500"`
	Method           string    `json:"method" gorm:"size:20"`
	PostDeployScript string    `json:"post_deploy_script" gorm:"type:text"`
	SortOrder        int       `json:"sort_order" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
