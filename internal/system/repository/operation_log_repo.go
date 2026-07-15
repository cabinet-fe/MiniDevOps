package repository

import (
	"time"

	"bedrock/internal/system/model"

	"gorm.io/gorm"
)

type OperationLogRepository struct {
	db *gorm.DB
}

func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

func (r *OperationLogRepository) Create(log *model.OperationLog) error {
	return r.db.Create(log).Error
}

type OperationLogFilters struct {
	Page         int
	PageSize     int
	UserID       *uint
	Action       string
	ResourceType string
	From         *time.Time
	To           *time.Time
}

func (r *OperationLogRepository) List(f OperationLogFilters) ([]model.OperationLog, int64, error) {
	q := r.db.Model(&model.OperationLog{})
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.ResourceType != "" {
		q = q.Where("resource_type = ?", f.ResourceType)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.OperationLog
	err := q.Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Order("id DESC").Find(&items).Error
	return items, total, err
}
