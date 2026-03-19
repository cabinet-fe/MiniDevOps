package model

import "time"

type EnvVar struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	EnvironmentID uint      `json:"environment_id" gorm:"index;not null"`
	Key           string    `json:"key" gorm:"size:200;not null"`
	Value         string    `json:"value" gorm:"type:text"`
	IsSecret      bool      `json:"is_secret" gorm:"default:false"`
	Masked        bool      `json:"masked,omitempty" gorm:"-"`
	HasValue      bool      `json:"has_value,omitempty" gorm:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type VarGroup struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Description string         `json:"description" gorm:"size:500"`
	Items       []VarGroupItem `json:"items,omitempty" gorm:"foreignKey:VarGroupID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type VarGroupItem struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	VarGroupID uint      `json:"var_group_id" gorm:"index;not null"`
	Key        string    `json:"key" gorm:"size:200;not null"`
	Value      string    `json:"value" gorm:"type:text"`
	IsSecret   bool      `json:"is_secret" gorm:"default:false"`
	Masked     bool      `json:"masked,omitempty" gorm:"-"`
	HasValue   bool      `json:"has_value,omitempty" gorm:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type EnvironmentVarGroup struct {
	EnvironmentID uint      `json:"environment_id" gorm:"primaryKey"`
	VarGroupID    uint      `json:"var_group_id" gorm:"primaryKey"`
	CreatedAt     time.Time `json:"created_at"`
}
