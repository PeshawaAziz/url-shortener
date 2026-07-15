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
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index:idx_user_tenant"`
	Email        string    `gorm:"type:citext;not null"` // requires citext extension
	DisplayName  string
	AuthProvider AuthProvider `gorm:"not null;default:'email'"`
	AuthSubject  string       `gorm:"not null"`
	CreatedAt    time.Time    `gorm:"not null;default:now()"`
	UpdatedAt    time.Time    `gorm:"not null;default:now()"`
	// Handled in migration: UNIQUE (tenant_id, email)
}

func NewUser(tenantID uuid.UUID, email, displayName string, provider AuthProvider, subject string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		DisplayName:  displayName,
		AuthProvider: provider,
		AuthSubject:  subject,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
