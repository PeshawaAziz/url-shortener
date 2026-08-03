package tenant

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanEnterprise Plan = "enterprise"
)

type Tenant struct {
	ID        uuid.UUID
	Name      string
	Slug      string // URL‑friendly unique identifier
	Plan      Plan
	Features  json.RawMessage // JSON map of feature flags
	Config    json.RawMessage // arbitrary tenant‑specific configuration
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTenant(name, slug string, plan Plan, features json.RawMessage, config json.RawMessage) *Tenant {
	now := time.Now()
	if features == nil {
		features = json.RawMessage(`{}`)
	}
	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		Plan:      plan,
		Features:  features,
		Config:    config,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Tenant) HasFeature(feature string) bool {
	var m map[string]bool
	if err := json.Unmarshal(t.Features, &m); err != nil {
		return false
	}
	return m[feature]
}
