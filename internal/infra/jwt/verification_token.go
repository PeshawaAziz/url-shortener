package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type VerificationTokenService struct {
	secret string
	ttl    time.Duration
}

func NewVerificationTokenService(secret string, ttl time.Duration) *VerificationTokenService {
	return &VerificationTokenService{secret: secret, ttl: ttl}
}

type verificationClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

func (s *VerificationTokenService) Issue(userID uuid.UUID, email string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(s.ttl)
	claims := verificationClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Email: email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, exp, nil
}

func (s *VerificationTokenService) Validate(tokenString string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &verificationClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}
	claims, ok := token.Claims.(*verificationClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", err
	}
	return userID, claims.Email, nil
}
