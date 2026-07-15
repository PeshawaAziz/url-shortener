package tag

import "github.com/google/uuid"

type Tag struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name     string    `gorm:"not null"`
	TenantID uuid.UUID `gorm:"type:uuid;not null"`
	// GORM many2many relation would be defined in URL, not here.
}

func NewTag(tenantID uuid.UUID, name string) *Tag {
	return &Tag{
		ID:       uuid.New(),
		Name:     name,
		TenantID: tenantID,
	}
}
