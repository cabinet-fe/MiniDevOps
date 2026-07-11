package model

import "time"

// Agent 用户配置的智能体（提示词 + CLI 代理 + 项目范围）
type Agent struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"size:100;not null"`
	Prompt     string    `json:"prompt" gorm:"type:text"`
	ProxyKey   string    `json:"proxy_key" gorm:"size:50;not null"` // opencode | claude | reasonix
	Enabled    bool      `json:"enabled" gorm:"default:true"`
	ProjectIDs []uint    `json:"project_ids" gorm:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AgentProject 智能体与项目的多对多范围
type AgentProject struct {
	AgentID   uint      `json:"agent_id" gorm:"primaryKey"`
	ProjectID uint      `json:"project_id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}

// EnvironmentAgent 环境挂载的智能体（按 sort_order 顺序执行）
type EnvironmentAgent struct {
	EnvironmentID uint      `json:"environment_id" gorm:"primaryKey"`
	AgentID       uint      `json:"agent_id" gorm:"primaryKey"`
	SortOrder     int       `json:"sort_order" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`
}
