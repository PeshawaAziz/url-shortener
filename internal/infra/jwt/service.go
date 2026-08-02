package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Config struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	PrivateKey      *rsa.PrivateKey
	PublicKey       *rsa.PublicKey
}

type JWTTokenService struct {
	config            Config
	refreshTokenStore auth.RefreshTokenStore
}

func NewJWTTokenService(cfg Config, store auth.RefreshTokenStore) *JWTTokenService {
	return &JWTTokenService{
		config:            cfg,
		refreshTokenStore: store,
	}
}

func (s *JWTTokenService) GenerateAccessToken(ctx context.Context, userID, tenantID uuid.UUID) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(s.config.AccessTokenTTL)

	claims := jwt.MapClaims{
		"uid": userID.String(),
		"tid": tenantID.String(),
		"iat": now.Unix(),
		"exp": exp.Unix(),
		"sub": userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.config.PrivateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, exp, nil
}

func (s *JWTTokenService) GenerateRefreshToken(ctx context.Context, userID, tenantID uuid.UUID) (string, string, time.Time, error) {
	rawToken, err := randomBytes()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	familyID := uuid.New().String()
	hashedToken := hashToken(rawToken)
	exp := time.Now().Add(s.config.RefreshTokenTTL)

	meta := auth.RefreshTokenMetadata{
		UserID:    userID,
		TenantID:  tenantID,
		FamilyID:  familyID,
		ExpiresAt: exp,
		Used:      false,
	}

	if err := s.refreshTokenStore.Save(ctx, hashedToken, meta); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to store refresh token: %w", err)
	}
	return rawToken, familyID, exp, nil
}

func (s *JWTTokenService) ValidateAccessToken(ctx context.Context, tokenString string) (*auth.AccessTokenClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.PublicKey, nil
	}

	parsed, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, auth.ErrInvalidAccessToken
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, auth.ErrInvalidAccessToken
	}

	uidStr, _ := claims["uid"].(string)
	tidStr, _ := claims["tid"].(string)
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return nil, auth.ErrInvalidAccessToken
	}
	tid, err := uuid.Parse(tidStr)
	if err != nil {
		return nil, auth.ErrInvalidAccessToken
	}
	iat, _ := claims["iat"].(float64)
	exp, _ := claims["exp"].(float64)

	return &auth.AccessTokenClaims{
		UserID:    uid,
		TenantID:  tid,
		IssuedAt:  int64(iat),
		ExpiresAt: int64(exp),
		Subject:   uidStr,
	}, nil
}

func (s *JWTTokenService) RotateRefreshToken(ctx context.Context, rawRefreshToken string) (string, string, time.Time, error) {
	hashedOld := hashToken(rawRefreshToken)
	oldMeta, err := s.refreshTokenStore.Get(ctx, hashedOld)
	if err != nil {
		return "", "", time.Time{}, auth.ErrInvalidRefreshToken
	}
	if time.Now().After(oldMeta.ExpiresAt) {
		return "", "", time.Time{}, auth.ErrInvalidRefreshToken
	}
	if oldMeta.Used {
		_ = s.refreshTokenStore.RevokeFamily(ctx, oldMeta.FamilyID)
		return "", "", time.Time{}, auth.ErrRefreshTokenReused
	}

	if err := s.refreshTokenStore.MarkUsed(ctx, hashedOld); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to mark old token used: %w", err)
	}

	newRawToken, err := randomBytes()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	newHashed := hashToken(newRawToken)

	newAccessToken, _, err := s.GenerateAccessToken(ctx, oldMeta.UserID, oldMeta.TenantID)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	newMeta := auth.RefreshTokenMetadata{
		UserID:    oldMeta.UserID,
		TenantID:  oldMeta.TenantID,
		FamilyID:  oldMeta.FamilyID,
		ExpiresAt: time.Now().Add(s.config.RefreshTokenTTL),
		Used:      false,
	}

	if err := s.refreshTokenStore.Save(ctx, newHashed, newMeta); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	return newAccessToken, newRawToken, newMeta.ExpiresAt, nil
}

func (s *JWTTokenService) RevokeRefreshTokenFamily(ctx context.Context, familyID string) error {
	return s.refreshTokenStore.RevokeFamily(ctx, familyID)
}

func randomBytes() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *JWTTokenService) RevokeRefreshToken(ctx context.Context, rawRefreshToken string) error {
	hashed := hashToken(rawRefreshToken)
	meta, err := s.refreshTokenStore.Get(ctx, hashed)
	if err != nil {
		if err == auth.ErrTokenNotFound {
			return nil
		}
		return err
	}
	return s.refreshTokenStore.RevokeFamily(ctx, meta.FamilyID)
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
