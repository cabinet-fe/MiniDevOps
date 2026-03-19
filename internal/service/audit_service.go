package service

import (
	"time"

	"buildflow/internal/model"
	"buildflow/internal/repository"
)

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(userID uint, action, resourceType string, resourceID uint, details, ipAddress string) error {
	entry := &model.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
	}
	return s.repo.Create(entry)
}

// AuditListFilters holds filters for listing audit logs.
type AuditListFilters struct {
	Action       string
	UserID       *uint
	ResourceType string
	From         time.Time
	To           time.Time
	Page         int
	PageSize     int
}

func (s *AuditService) List(filters *AuditListFilters) ([]model.AuditLog, int64, error) {
	var repoFilters *repository.AuditListFilters
	if filters != nil {
		repoFilters = &repository.AuditListFilters{
			Action:       filters.Action,
			UserID:       filters.UserID,
			ResourceType: filters.ResourceType,
			FromDate:     filters.From,
			ToDate:       filters.To,
			Page:         filters.Page,
			PageSize:     filters.PageSize,
		}
		if repoFilters.Page < 1 {
			repoFilters.Page = 1
		}
		if repoFilters.PageSize < 1 {
			repoFilters.PageSize = 20
		}
	}
	return s.repo.List(repoFilters)
}
