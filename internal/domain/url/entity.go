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
	StateDeleted  URLState = "deleted"
)

type RedirectType string

const (
	RedirectTypePermanent RedirectType = "permenant"
	RedirectTypeTemporary RedirectType = "temporary"
)

type URL struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	UserID      uuid.UUID
	Slug        Slug
	OriginalURL OriginalURL

	ExpiresAt      *time.Time
	ClickCap       *int64
	ClickCount     int64
	PasswordHash   string       // Empty means no password
	RedirectType   RedirectType // "permanent" (301) or "temporary" (302)
	RoutingConfig  RoutingConfig
	RateLimitPerHr *int

	State     URLState
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewURL(tenantID, userID uuid.UUID, slug Slug, dest OriginalURL) *URL {
	now := time.Now()
	return &URL{
		ID:           uuid.New(),
		TenantID:     tenantID,
		UserID:       userID,
		Slug:         slug,
		OriginalURL:  dest,
		RedirectType: RedirectTypeTemporary,
		State:        StateActive,
		CreatedAt:    now,
		UpdatedAt:    now,
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

func (u *URL) IsDeleted() bool {
	return u.State == StateDeleted || u.DeletedAt != nil
}

func (u *URL) Pause() error {
	if u.State != StateActive {
		return errors.New("only active urls can be paused")
	}
	u.State = StatePaused
	u.UpdatedAt = time.Now()
	return nil
}

func (u *URL) Resume() error {
	if u.State != StatePaused {
		return errors.New("only paused urls can be resumed")
	}
	u.State = StateActive
	u.UpdatedAt = time.Now()
	return nil
}

func (u *URL) Archive() error {
	if u.State == StateDeleted {
		return errors.New("cannot archive a deleted url")
	}
	u.State = StateArchived
	u.UpdatedAt = time.Now()
	return nil
}

func (u *URL) SoftDelete() error {
	if u.State == StateDeleted {
		return errors.New("already deleted")
	}
	now := time.Now()
	u.State = StateDeleted
	u.DeletedAt = &now
	u.UpdatedAt = now
	return nil
}

func (u *URL) IncrementClick() {
	u.ClickCount++
	u.UpdatedAt = time.Now()
}

func (u *URL) IsPasswordProtected() bool {
	return u.PasswordHash != ""
}

func (u *URL) HasRateLimit() bool {
	return u.RateLimitPerHr != nil && *u.RateLimitPerHr > 0
}
