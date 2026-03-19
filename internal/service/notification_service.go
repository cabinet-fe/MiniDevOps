package service

import (
	"buildflow/internal/model"
	"buildflow/internal/repository"
)

type NotificationService struct {
	repo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) CreateNotification(userID uint, notifType, title, message string, buildID *uint) error {
	n := &model.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		BuildID: buildID,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) ListByUser(userID uint, page, pageSize int) ([]model.Notification, int64, error) {
	return s.repo.ListByUser(userID, page, pageSize)
}

func (s *NotificationService) MarkRead(id, userID uint) error {
	return s.repo.MarkRead(id, userID)
}

func (s *NotificationService) MarkAllRead(userID uint) error {
	return s.repo.MarkAllRead(userID)
}

func (s *NotificationService) CountUnread(userID uint) (int64, error) {
	return s.repo.CountUnread(userID)
}
