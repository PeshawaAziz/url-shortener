package url

import "time"

type URL struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string     `gorm:"type:uuid;not null;index" json:"user_id"`
	OriginalURL  string     `gorm:"not null" json:"original_url"`
	ShortCode    string     `gorm:"uniqueIndex;not null;size:20" json:"short_code"`
	CustomAlias  *string    `gorm:"uniqueIndex;size:50" json:"custom_alias,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	MaxVisits    *int       `json:"max_visits,omitempty"`
	VisitCount   int        `gorm:"not null;default:0" json:"visit_count"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	PasswordHash *string    `gorm:"size:60" json:"-"` // future feature
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
