package model

import "time"

// BuildDistribution records one execution of one distribution for a build.
// Unique (build_id, distribution_id): one row per target per build; updated on re-distribute.
type BuildDistribution struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	BuildID        uint       `json:"build_id" gorm:"not null;uniqueIndex:idx_build_distribution_pair"`
	DistributionID uint       `json:"distribution_id" gorm:"not null;uniqueIndex:idx_build_distribution_pair"`
	Status         string     `json:"status" gorm:"size:20;not null;default:pending"`
	ErrorMessage   string     `json:"error_message" gorm:"type:text"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
