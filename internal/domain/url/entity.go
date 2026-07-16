package url

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type URLState string

const (
	StateActive   URLState = "active"
	StatePaused   URLState = "paused"
	StateExpired  URLState = "expired"
	StateArchived URLState = "archived"
)

type URL struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	UserID      uuid.UUID
	Slug        Slug
	OriginalURL OriginalURL

	ExpiresAt    *time.Time
	ClickCap     *int64
	ClickCount   int64
	PasswordHash string // Empty means no password

	State        URLState
	RedirectType string `json:"redirect_type"` // "permanent" (301) or "temporary" (302)
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewURL(tenantID, userID uuid.UUID, slug Slug, dest OriginalURL) *URL {
	now := time.Now()
	return &URL{
		ID:          uuid.New(),
		TenantID:    tenantID,
		UserID:      userID,
		Slug:        slug,
		OriginalURL: dest,
		State:       StateActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (u *URL) CanRedirect(now time.Time) bool {
	if u.State != StateActive {
		return false
	}
	if u.ExpiresAt != nil && now.After(*u.ExpiresAt) {
		return false
	}
	if u.ClickCap != nil && u.ClickCount >= *u.ClickCap {
		return false
	}
	return true
}

func (u *URL) Pause() error {
	if u.State != StateActive {
		return errors.New("only active urls can be paused")
	}
	u.State = StatePaused
	u.UpdatedAt = time.Now()
	return nil
}

func (u *URL) IncrementClick() {
	u.ClickCount++
	u.UpdatedAt = time.Now()
}
