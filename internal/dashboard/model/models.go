package model

import "time"

// Layout is a per-user dashboard card configuration. CardsJSON is kept as a
// JSON array so a new card type does not require a schema migration.
type Layout struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex"`
	CardsJSON string    `json:"-" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Layout) TableName() string { return "dashboard_layouts" }

// CardLayout controls one known dashboard card. Card IDs are server-defined;
// clients cannot register arbitrary cards through the layout endpoint.
type CardLayout struct {
	ID      string `json:"id"`
	Visible bool   `json:"visible"`
	Order   int    `json:"order"`
}

// LayoutResponse is intentionally stable for the dashboard editor.
type LayoutResponse struct {
	Cards []CardLayout `json:"cards"`
}

type BuildSummary struct {
	Running     int64       `json:"running"`
	Queued      int64       `json:"queued"`
	SuccessRate float64     `json:"success_rate"`
	Recent      []RecentRun `json:"recent"`
}

type RecentRun struct {
	ID          uint      `json:"id"`
	BuildJobID  uint      `json:"build_job_id"`
	BuildNumber int       `json:"build_number"`
	Status      string    `json:"status"`
	Branch      string    `json:"branch"`
	CreatedAt   time.Time `json:"created_at"`
}

type SystemInfo struct {
	Version   string    `json:"version"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	Runtime   string    `json:"runtime"`
	Hostname  string    `json:"hostname"`
	StartTime time.Time `json:"start_time"`
}

type DiskStatus struct {
	Path       string  `json:"path"`
	TotalBytes uint64  `json:"total_bytes"`
	FreeBytes  uint64  `json:"free_bytes"`
	UsedPct    float64 `json:"used_percent"`
}

type SystemStatus struct {
	CPUUsagePercent    float64      `json:"cpu_usage_percent"`
	MemoryUsedBytes    uint64       `json:"memory_used_bytes"`
	MemoryTotalBytes   uint64       `json:"memory_total_bytes"`
	MemoryUsagePercent float64      `json:"memory_usage_percent"`
	Health             string       `json:"health"`
	Directories        []DiskStatus `json:"directories"`
	CollectedAt        time.Time    `json:"collected_at"`
}
