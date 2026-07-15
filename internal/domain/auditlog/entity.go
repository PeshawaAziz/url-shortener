package auditlog

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         int64           `gorm:"primaryKey;autoIncrement"`
	TenantID   uuid.UUID       `gorm:"type:uuid;not null;index"`
	UserID     *uuid.UUID      `gorm:"type:uuid"`
	Action     string          `gorm:"not null"`
	EntityType string          `gorm:"not null"`
	EntityID   string          `gorm:"not null"`
	OldValues  json.RawMessage `gorm:"type:jsonb"`
	NewValues  json.RawMessage `gorm:"type:jsonb"`
	CreatedAt  time.Time       `gorm:"not null;default:now()"`
}

func NewAuditLog(tenantID uuid.UUID, userID *uuid.UUID, action, entityType, entityID string, oldValues, newValues json.RawMessage) *AuditLog {
	return &AuditLog{
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValues:  oldValues,
		NewValues:  newValues,
		CreatedAt:  time.Now(),
	}
}
