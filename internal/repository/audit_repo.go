package repository

import (
	"time"

	"buildflow/internal/model"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(audit *model.AuditLog) error {
	return r.db.Create(audit).Error
}

func (r *AuditRepository) List(filters *AuditListFilters) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	query := r.db.Model(&model.AuditLog{})
	if filters != nil {
		if filters.Action != "" {
			query = query.Where("action = ?", filters.Action)
		}
		if filters.UserID != nil {
			query = query.Where("user_id = ?", *filters.UserID)
		}
		if filters.ResourceType != "" {
			query = query.Where("resource_type = ?", filters.ResourceType)
		}
		if !filters.FromDate.IsZero() {
			query = query.Where("created_at >= ?", filters.FromDate)
		}
		if !filters.ToDate.IsZero() {
			query = query.Where("created_at <= ?", filters.ToDate)
		}
	}
	query.Count(&total)
	page, pageSize := 1, 20
	if filters != nil {
		page, pageSize = filters.Page, filters.PageSize
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 20
		}
	}
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

type AuditListFilters struct {
	Action       string
	UserID       *uint
	ResourceType string
	FromDate     time.Time
	ToDate       time.Time
	Page         int
	PageSize     int
}
