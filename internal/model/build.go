package model

import "time"

type Build struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	ProjectID     uint       `json:"project_id" gorm:"uniqueIndex:idx_proj_build_num;not null"`
	EnvironmentID uint       `json:"environment_id" gorm:"not null"`
	BuildNumber   int        `json:"build_number" gorm:"uniqueIndex:idx_proj_build_num;not null"`
	Status        string     `json:"status" gorm:"size:20;not null;default:pending"`
	CurrentStage  string     `json:"current_stage" gorm:"size:20;not null;default:pending"`
	TriggerType   string     `json:"trigger_type" gorm:"size:20"`
	TriggeredBy   uint       `json:"triggered_by"`
	Branch        string     `json:"branch" gorm:"size:200"`
	CommitHash    string     `json:"commit_hash" gorm:"size:40"`
	CommitMessage string     `json:"commit_message" gorm:"size:500"`
	LogPath       string     `json:"log_path" gorm:"size:500"`
	ArtifactPath  string     `json:"artifact_path" gorm:"size:500"`
	DurationMs             int64      `json:"duration_ms"`
	ErrorMessage           string     `json:"error_message" gorm:"type:text"`
	DistributionSummary    string     `json:"distribution_summary" gorm:"size:30"`
	RedistributeFilterJSON string     `json:"redistribute_filter_json,omitempty" gorm:"type:text"`
	StartedAt              *time.Time `json:"started_at"`
	FinishedAt             *time.Time `json:"finished_at"`
	CreatedAt              time.Time  `json:"created_at"`
}
