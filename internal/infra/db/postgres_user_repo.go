package db

import (
	"context"
	"errors"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userModel struct {
	ID                  string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID            string     `gorm:"column:tenant_id;not null;index:idx_user_tenant"`
	Email               string     `gorm:"column:email;type:citext;not null;uniqueIndex:idx_user_tenant_email"`
	DisplayName         string     `gorm:"column:display_name"`
	AuthProvider        string     `gorm:"column:auth_provider;not null;default:'email'"`
	AuthSubject         string     `gorm:"column:auth_subject;not null"`
	PasswordHash        *string    `gorm:"column:password_hash"`
	IsEmailVerified     bool       `gorm:"column:is_email_verified;not null;default:false"`
	EmailVerifiedAt     *time.Time `gorm:"column:email_verified_at"`
	FailedLoginAttempts int        `gorm:"column:failed_login_attempts;not null;default:0"`
	LockedUntil         *time.Time `gorm:"column:locked_until"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (userModel) TableName() string {
	return "users"
}

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (m *userModel) toDomain() *user.User {
	return &user.User{
		ID:                  uuid.MustParse(m.ID),
		TenantID:            uuid.MustParse(m.TenantID),
		Email:               m.Email,
		DisplayName:         m.DisplayName,
		AuthProvider:        user.AuthProvider(m.AuthProvider),
		AuthSubject:         m.AuthSubject,
		PasswordHash:        m.PasswordHash,
		IsEmailVerified:     m.IsEmailVerified,
		EmailVerifiedAt:     m.EmailVerifiedAt,
		FailedLoginAttempts: m.FailedLoginAttempts,
		LockedUntil:         m.LockedUntil,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func fromDomain(u *user.User) *userModel {
	return &userModel{
		ID:                  u.ID.String(),
		TenantID:            u.TenantID.String(),
		Email:               u.Email,
		DisplayName:         u.DisplayName,
		AuthProvider:        string(u.AuthProvider),
		AuthSubject:         u.AuthSubject,
		PasswordHash:        u.PasswordHash,
		IsEmailVerified:     u.IsEmailVerified,
		EmailVerifiedAt:     u.EmailVerifiedAt,
		FailedLoginAttempts: u.FailedLoginAttempts,
		LockedUntil:         u.LockedUntil,
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
	}
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*user.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND email = ?", tenantID.String(), email).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id.String()).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *user.User) error {
	m := fromDomain(u)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil && isDuplicateKeyError(err) {
		return user.ErrEmailAlreadyTaken
	}
	return err
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *user.User) error {
	m := fromDomain(u)
	return r.db.WithContext(ctx).Save(m).Error
}
