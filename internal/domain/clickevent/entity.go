package clickevent

import (
	"time"

	"github.com/google/uuid"
)

type ClickEvent struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	URLID      uuid.UUID `gorm:"type:uuid;not null;index:idx_clicks_url_time,priority:1"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index:idx_clicks_tenant_time,priority:1"`
	ClickedAt  time.Time `gorm:"not null;index:idx_clicks_url_time,priority:2;index:idx_clicks_tenant_time,priority:2"`
	IP         string
	UserAgent  string
	Referer    string
	Country    string
	DeviceType string
}

func (ClickEvent) TableName() string {
	return "click_events" // partitioned table
}

func NewClickEvent(urlID, tenantID uuid.UUID, ip, userAgent, referer, country, deviceType string) *ClickEvent {
	return &ClickEvent{
		URLID:      urlID,
		TenantID:   tenantID,
		ClickedAt:  time.Now(),
		IP:         ip,
		UserAgent:  userAgent,
		Referer:    referer,
		Country:    country,
		DeviceType: deviceType,
	}
}
