package db

import (
	"context"
	"errors"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/url"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type urlModel struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID     uuid.UUID  `gorm:"column:tenant_id;not null"`
	UserID       uuid.UUID  `gorm:"column:user_id;not null"`
	Slug         string     `gorm:"column:slug;not null"`
	OriginalURL  string     `gorm:"column:original_url;not null"`
	ExpiresAt    *time.Time `gorm:"column:expires_at"`
	ClickCap     *int64     `gorm:"column:click_cap"`
	ClickCount   int64      `gorm:"column:click_count;default:0"`
	PasswordHash string     `gorm:"column:password_hash"`
	State        string     `gorm:"column:state;not null;default:'active'"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (urlModel) TableName() string { return "urls" }

type PostgresURLRepository struct {
	db *gorm.DB
}

func NewPostgresURLRepository(db *gorm.DB) *PostgresURLRepository {
	return &PostgresURLRepository{db: db}
}

func toModel(u *url.URL) *urlModel {
	return &urlModel{
		ID:           u.ID,
		TenantID:     u.TenantID,
		UserID:       u.UserID,
		Slug:         string(u.Slug),
		OriginalURL:  string(u.OriginalURL),
		ExpiresAt:    u.ExpiresAt,
		ClickCap:     u.ClickCap,
		ClickCount:   u.ClickCount,
		PasswordHash: u.PasswordHash,
		State:        string(u.State),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func toDomain(m *urlModel) *url.URL {
	return &url.URL{
		ID:           m.ID,
		TenantID:     m.TenantID,
		UserID:       m.UserID,
		Slug:         url.Slug(m.Slug),
		OriginalURL:  url.OriginalURL(m.OriginalURL),
		ExpiresAt:    m.ExpiresAt,
		ClickCap:     m.ClickCap,
		ClickCount:   m.ClickCount,
		PasswordHash: m.PasswordHash,
		State:        url.URLState(m.State),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (r *PostgresURLRepository) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*url.URL, error) {
	var m urlModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).
		First(&m).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, url.ErrURLNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *PostgresURLRepository) Save(ctx context.Context, u *url.URL) error {
	model := toModel(u)
	err := r.db.WithContext(ctx).Create(model).Error

	// Check for Postgres unique violation (SQLSTATE 23505)
	if err != nil && err.Error() != "" && contains(err.Error(), "23505") {
		return url.ErrSlugConflict
	}
	return err
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
