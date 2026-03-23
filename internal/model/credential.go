package model

import "time"

type Credential struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null;uniqueIndex:idx_credential_name_creator"`
	Type        string    `json:"type" gorm:"size:20;not null"`
	Username    string    `json:"username" gorm:"size:200"`
	Password    string    `json:"-" gorm:"size:1000"`
	Description string    `json:"description" gorm:"size:500"`
	CreatedBy   uint      `json:"created_by" gorm:"not null;uniqueIndex:idx_credential_name_creator"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// HasSecret indicates whether secret value exists and is only used in API payload.
	HasSecret bool `json:"has_secret" gorm:"-"`

	// CreatorName is a derived field for display in list APIs.
	CreatorName string `json:"creator_name,omitempty" gorm:"-"`
}
