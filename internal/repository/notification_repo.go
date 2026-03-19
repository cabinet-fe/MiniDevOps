package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) ListByUser(userID uint, page, pageSize int) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64
	r.db.Model(&model.Notification{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.Where("user_id = ?", userID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, total, err
}

func (r *NotificationRepository) MarkRead(id uint, userID uint) error {
	return r.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllRead(userID uint) error {
	return r.db.Model(&model.Notification{}).
		Where("user_id = ?", userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
