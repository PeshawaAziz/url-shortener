package serviceaccount

import (
	"time"

	"github.com/google/uuid"
)

type Scope string

const (
	ScopeReadLinks     Scope = "read:links"
	ScopeWriteLinks    Scope = "write:links"
	ScopeReadAnalytics Scope = "read:analytics"
)

type ServiceAccount struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	CreatedBy   *uuid.UUID // admin who created it

	KeyHash string

	Scopes []Scope

	IPAllowlist  []string
	RateLimitRPM int // 0 = unlimited

	Active bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // soft delete
}

func NewServiceAccount(tenantID uuid.UUID, name, description string, createdBy *uuid.UUID, keyHash string, scopes []Scope) *ServiceAccount {
	now := time.Now()
	if scopes == nil {
		scopes = []Scope{}
	}
	return &ServiceAccount{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		KeyHash:     keyHash,
		Scopes:      scopes,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (sa *ServiceAccount) IsActive() bool {
	return sa.Active && sa.DeletedAt == nil
}

func (sa *ServiceAccount) HasScope(s Scope) bool {
	for _, sc := range sa.Scopes {
		if sc == s {
			return true
		}
	}
	return false
}

func (sa *ServiceAccount) SoftDelete() {
	now := time.Now()
	sa.DeletedAt = &now
	sa.UpdatedAt = now
}
