package db

import (
	"context"
	"errors"
	"strings"
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
	RedirectType string     `gorm:"column:redirect_type;not null;default:'temporary'"`
	State        string     `gorm:"column:state;not null;default:'active'"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (urlModel) TableName() string {
	return "urls"
}

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
		RedirectType: string(u.RedirectType),
		State:        string(u.State),
		DeletedAt:    u.DeletedAt,
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
		RedirectType: url.RedirectType(m.RedirectType),
		State:        url.URLState(m.State),
		DeletedAt:    m.DeletedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (r *PostgresURLRepository) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*url.URL, error) {
	var m urlModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND slug = ? AND state = ?", tenantID, slug, "active").
		First(&m).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, url.ErrURLNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *PostgresURLRepository) GetDeletedBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*url.URL, error) {
	var m urlModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND slug = ? AND state = ?", tenantID, slug, "deleted").
		First(&m).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, url.ErrURLNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *PostgresURLRepository) GetByID(ctx context.Context, id uuid.UUID) (*url.URL, error) {
	var m urlModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error

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

	if err != nil && isDuplicateKeyError(err) {
		return url.ErrSlugConflict
	}
	return err
}

func (r *PostgresURLRepository) Update(ctx context.Context, u *url.URL) error {
	model := toModel(u)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PostgresURLRepository) ListActiveForSweep(ctx context.Context, limit int, offset int) ([]*url.URL, error) {
	var models []urlModel
	now := time.Now()

	err := r.db.WithContext(ctx).
		Where("state = ?", "active").
		Where("expires_at IS NOT NULL AND expires_at < ?", now).
		Or("click_cap IS NOT NULL AND click_count >= click_cap").
		Limit(limit).
		Offset(offset).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	var urls []*url.URL
	for _, m := range models {
		urls = append(urls, toDomain(&m))
	}
	return urls, nil
}

// isDuplicateKeyError checks for Postgres SQLSTATE 23505
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505")
}
