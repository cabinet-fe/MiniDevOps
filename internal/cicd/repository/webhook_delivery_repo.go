package repository

import (
	"time"

	"gorm.io/gorm"
)

// WebhookDelivery records processed deliveries for idempotency.
type WebhookDelivery struct {
	ID          uint      `gorm:"primaryKey"`
	BuildJobID  uint      `gorm:"uniqueIndex:idx_wh_delivery;not null"`
	DeliveryKey string    `gorm:"size:200;uniqueIndex:idx_wh_delivery;not null"`
	CreatedAt   time.Time `gorm:""`
}

func (WebhookDelivery) TableName() string { return "webhook_deliveries" }

type WebhookDeliveryRepository struct{ db *gorm.DB }

func NewWebhookDeliveryRepository(db *gorm.DB) *WebhookDeliveryRepository {
	return &WebhookDeliveryRepository{db: db}
}

// TryInsert returns true if this delivery is new; false if duplicate.
func (r *WebhookDeliveryRepository) TryInsert(buildJobID uint, deliveryKey string) (bool, error) {
	row := &WebhookDelivery{BuildJobID: buildJobID, DeliveryKey: deliveryKey}
	err := r.db.Create(row).Error
	if err != nil {
		// Unique violation → duplicate
		if isUniqueViolation(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return containsAny(msg, "UNIQUE", "unique", "Duplicate", "duplicate")
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if len(p) > 0 && (len(s) >= len(p)) {
			for i := 0; i+len(p) <= len(s); i++ {
				if s[i:i+len(p)] == p {
					return true
				}
			}
		}
	}
	return false
}
