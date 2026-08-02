package auth

import "errors"

var (
	ErrInvalidAccessToken  = errors.New("invalid or expired access token")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrRefreshTokenReused  = errors.New("refresh token reuse detected – family revoked")
	ErrTokenNotFound       = errors.New("refresh token not found in store")
)
