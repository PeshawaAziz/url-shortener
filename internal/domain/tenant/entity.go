package tenant

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string          `gorm:"not null;uniqueIndex"`
	Slug      string          `gorm:"not null;uniqueIndex"`
	Config    json.RawMessage `gorm:"type:jsonb"`
	CreatedAt time.Time       `gorm:"not null;default:now()"`
	UpdatedAt time.Time       `gorm:"not null;default:now()"`
}

func NewTenant(name, slug string, config json.RawMessage) *Tenant {
	now := time.Now()
	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		Config:    config,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
