package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenService interface {
	// GenerateAccessToken creates a short‑lived access token (JWT).
	GenerateAccessToken(ctx context.Context, userID, tenantID uuid.UUID) (string, time.Time, error)

	// GenerateRefreshToken creates a refresh token and persists its metadata.
	// Returns the raw refresh token (to be set as httpOnly cookie) and the token family ID.
	GenerateRefreshToken(ctx context.Context, userID, tenantID uuid.UUID) (string, string, time.Time, error)

	// ValidateAccessToken parses and validates a JWT access token.
	// Returns claims if valid, or an error.
	ValidateAccessToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error)

	// RotateRefreshToken rotates the given refresh token. It marks the old token as used,
	// checks for reuse, and issues a new refresh token. Returns the new raw token and family ID.
	// If reuse is detected, the entire family is revoked and an error is returned.
	RotateRefreshToken(ctx context.Context, rawRefreshToken string) (string, string, time.Time, error)

	// RevokeRefreshTokenFamily revokes all refresh tokens belonging to a family.
	RevokeRefreshTokenFamily(ctx context.Context, familyID string) error
}

// AccessTokenClaims represents the JWT payload of an access token.
type AccessTokenClaims struct {
	UserID   uuid.UUID `json:"uid"`
	TenantID uuid.UUID `json:"tid"`
	// standard JWT claims
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Subject   string `json:"sub"` // userID as string
}

// RefreshTokenClaims is the payload stored in the refresh token (not JWT; it's an opaque token).
// This is only used internally for metadata storage.
type RefreshTokenMetadata struct {
	UserID    uuid.UUID
	TenantID  uuid.UUID
	FamilyID  string
	ExpiresAt time.Time
	Used      bool
}
