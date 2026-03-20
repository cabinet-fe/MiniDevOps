package model

import "time"

type Dictionary struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"size:100;not null"`
	Code        string     `json:"code" gorm:"uniqueIndex;size:100;not null"`
	Description string     `json:"description" gorm:"size:500"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Items       []DictItem `json:"items,omitempty" gorm:"foreignKey:DictionaryID"`
}

type DictItem struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	DictionaryID uint      `json:"dictionary_id" gorm:"index;not null"`
	Label        string    `json:"label" gorm:"size:200;not null"`
	Value        string    `json:"value" gorm:"size:200;not null"`
	SortOrder    int       `json:"sort_order" gorm:"default:0"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
