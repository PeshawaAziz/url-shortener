package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenService interface {
	GenerateAccessToken(ctx context.Context, userID, tenantID uuid.UUID) (string, time.Time, error)
	GenerateRefreshToken(ctx context.Context, userID, tenantID uuid.UUID) (string, string, time.Time, error)
	ValidateAccessToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error)
	RotateRefreshToken(ctx context.Context, rawRefreshToken string) (string, string, time.Time, error)
	RevokeRefreshTokenFamily(ctx context.Context, familyID string) error
	RevokeRefreshToken(ctx context.Context, rawRefreshToken string) error
}

type RefreshTokenStore interface {
	Save(ctx context.Context, hashedToken string, meta RefreshTokenMetadata) error
	Get(ctx context.Context, hashedToken string) (*RefreshTokenMetadata, error)
	MarkUsed(ctx context.Context, hashedToken string) error
	RevokeFamily(ctx context.Context, familyID string) error
}

type AccessTokenClaims struct {
	UserID    uuid.UUID `json:"uid"`
	TenantID  uuid.UUID `json:"tid"`
	IssuedAt  int64     `json:"iat"`
	ExpiresAt int64     `json:"exp"`
	Subject   string    `json:"sub"`
}

type RefreshTokenMetadata struct {
	UserID    uuid.UUID
	TenantID  uuid.UUID
	FamilyID  string
	ExpiresAt time.Time
	Used      bool
}

type VerificationTokenIssuer interface {
	Issue(userID uuid.UUID, email string) (string, time.Time, error)
	Validate(tokenString string) (userID uuid.UUID, email string, err error)
}
