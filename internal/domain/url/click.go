package url

import "time"

type Click struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	URLID     string    `gorm:"type:uuid;not null;index" json:"url_id"`
	ClickedAt time.Time `gorm:"not null;default:now()" json:"clicked_at"`
	Referer   *string   `json:"referer,omitempty"`
	UserAgent *string   `json:"user_agent,omitempty"`
	IPAddress *string   `json:"ip_address,omitempty"`
}
