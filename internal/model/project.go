package model

import "time"

type Project struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Name         string `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Description  string `json:"description" gorm:"size:500"`
	Tags         string `json:"tags" gorm:"size:500"`
	RepoURL      string `json:"repo_url" gorm:"size:500;not null"`
	RepoAuthType string `json:"repo_auth_type" gorm:"size:20;default:none"`
	CredentialID *uint  `json:"credential_id"`
	// Deprecated: kept for compatibility and migration fallback.
	RepoUsername string `json:"repo_username" gorm:"size:200"`
	// Deprecated: kept for compatibility and migration fallback.
	RepoPassword       string        `json:"-" gorm:"size:1000"`
	MaxArtifacts       int           `json:"max_artifacts" gorm:"default:5"`
	ArtifactFormat     string        `json:"artifact_format" gorm:"size:20;default:gzip"`
	WebhookSecret      string        `json:"webhook_secret" gorm:"size:64"`
	WebhookType        string        `json:"webhook_type" gorm:"size:20;default:auto"`
	WebhookRefPath     string        `json:"webhook_ref_path" gorm:"size:300"`
	WebhookCommitPath  string        `json:"webhook_commit_path" gorm:"size:300"`
	WebhookMessagePath string        `json:"webhook_message_path" gorm:"size:300"`
	CreatedBy          uint          `json:"created_by"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	Environments       []Environment `json:"environments,omitempty" gorm:"foreignKey:ProjectID"`
}
