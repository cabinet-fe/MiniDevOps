package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	KindAttachment = "attachment"
	KindDocImport  = "doc_import"
	KindSkillZIP   = "skill_zip"
	KindArtifact   = "artifact"
	KindOther      = "other"
)

// StorageObject is a content-addressed file held below the configured storage root.
// Path is always relative to that root and is never derived from client input.
type StorageObject struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Kind        string         `json:"kind" gorm:"size:32;not null;index"`
	SHA256      string         `json:"sha256" gorm:"size:64;not null;uniqueIndex"`
	Size        int64          `json:"size" gorm:"not null"`
	ContentType string         `json:"content_type" gorm:"size:200"`
	Path        string         `json:"path" gorm:"size:500;not null"`
	RefCount    int            `json:"ref_count" gorm:"not null;default:1"`
	CreatedBy   uint           `json:"created_by" gorm:"index"`
	PurgeAfter  *time.Time     `json:"purge_after,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (StorageObject) TableName() string { return "storage_objects" }
