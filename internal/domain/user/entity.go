package user

import (
	"time"

	"github.com/google/uuid"
)

type AuthProvider string

const (
	AuthProviderEmail  AuthProvider = "email"
	AuthProviderGoogle AuthProvider = "google"
)

type User struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Email       string
	DisplayName string

	AuthProvider AuthProvider
	AuthSubject  string // external ID (sub for OAuth, email for email)

	PasswordHash *string // nil if no password (for OAuth)

	IsEmailVerified bool
	EmailVerifiedAt *time.Time

	FailedLoginAttempts int
	LockedUntil         *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(tenantID uuid.UUID, email, displayName string, provider AuthProvider, subject string) *User {
	now := time.Now()
	if provider == AuthProviderEmail && subject == "" {
		subject = email
	}
	return &User{
		ID:              uuid.New(),
		TenantID:        tenantID,
		Email:           email,
		DisplayName:     displayName,
		AuthProvider:    provider,
		AuthSubject:     subject,
		IsEmailVerified: false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (u *User) HasPassword() bool {
	return u.PasswordHash != nil && *u.PasswordHash != ""
}

func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = &hash
	u.UpdatedAt = time.Now()
}

func (u *User) VerifyEmail() {
	now := time.Now()
	u.IsEmailVerified = true
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
}

func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && time.Now().Before(*u.LockedUntil)
}

func (u *User) RecordFailedLogin(maxAttempts int, lockDuration time.Duration) {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= maxAttempts {
		lockUntil := time.Now().Add(lockDuration)
		u.LockedUntil = &lockUntil
	}
	u.UpdatedAt = time.Now()
}

func (u *User) ClearLockout() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
}
