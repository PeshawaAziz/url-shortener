package url

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type URLState string

const (
	StateActive   URLState = "active"
	StateInactive URLState = "inactive"
	StateExpired  URLState = "expired"
)

type URL struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index:idx_tenant_slug_active,priority:1"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_user_urls"`
	Slug         string    `gorm:"not null;index:idx_tenant_slug_active,priority:2"`
	OriginalURL  string    `gorm:"not null"`
	Title        string
	Description  string
	ExpiresAt    *time.Time
	ClickCap     *int `gorm:"check:click_cap > 0"`
	PasswordHash *string
	State        URLState        `gorm:"not null;default:'active';index:idx_tenant_slug_active,priority:3,where:state='active'"`
	Metadata     json.RawMessage `gorm:"type:jsonb"`
	ClickCount   int             `gorm:"default:0"`
	CreatedAt    time.Time       `gorm:"not null;default:now()"`
	UpdatedAt    time.Time       `gorm:"not null;default:now()"`
	// Handled in migration:
	// - UNIQUE (tenant_id, slug)
	// - partial unique index: (tenant_id, slug) WHERE state = 'active'
}

func NewURL(tenantID, userID uuid.UUID, slug, originalURL string) *URL {
	now := time.Now()
	return &URL{
		ID:          uuid.New(),
		TenantID:    tenantID,
		UserID:      userID,
		Slug:        slug,
		OriginalURL: originalURL,
		State:       StateActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
