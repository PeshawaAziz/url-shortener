package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/shared" // for PasswordHasher
	"github.com/PeshawaAziz/url-shortener/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Config struct {
	SkipEmailVerification   bool
	VerificationTokenSecret string
	VerificationTokenTTL    time.Duration
	MaxLoginAttempts        int
	LockDuration            time.Duration
}

type UserAuthService struct {
	repo     user.UserRepository
	hasher   shared.PasswordHasher
	tokenSvc TokenService
	config   Config
}

func NewUserAuthService(repo user.UserRepository, hasher shared.PasswordHasher, tokenSvc TokenService, config Config) *UserAuthService {
	return &UserAuthService{
		repo:     repo,
		hasher:   hasher,
		tokenSvc: tokenSvc,
		config:   config,
	}
}

type RegisterInput struct {
	TenantID    uuid.UUID
	Email       string
	DisplayName string
	Password    string
}

type RegisterOutput struct {
	User              *user.User
	VerificationToken string
}

func (s *UserAuthService) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	existing, err := s.repo.FindByEmail(ctx, input.TenantID, input.Email)
	if err != nil && err != user.ErrUserNotFound {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if existing != nil {
		return nil, user.ErrEmailAlreadyTaken
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u := user.NewUser(input.TenantID, input.Email, input.DisplayName, user.AuthProviderEmail, input.Email)
	u.SetPasswordHash(hash)

	if s.config.SkipEmailVerification {
		u.VerifyEmail()
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	output := &RegisterOutput{User: u}

	if !s.config.SkipEmailVerification {
		token, err := s.generateVerificationToken(u.ID, u.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to generate verification token: %w", err)
		}
		output.VerificationToken = token
	}

	return output, nil
}

type LoginInput struct {
	TenantID uuid.UUID
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	User         *user.User
}

func (s *UserAuthService) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	u, err := s.repo.FindByEmail(ctx, input.TenantID, input.Email)
	if err != nil {
		return nil, user.ErrInvalidCredentials
	}

	if u.IsLocked() {
		return nil, user.ErrAccountLocked
	}

	if !s.config.SkipEmailVerification && !u.IsEmailVerified {
		return nil, user.ErrEmailNotVerified
	}

	if !u.HasPassword() || !s.hasher.Compare(*u.PasswordHash, input.Password) {
		u.RecordFailedLogin(s.config.MaxLoginAttempts, s.config.LockDuration)
		_ = s.repo.Update(ctx, u) // ignore error to not leak info
		return nil, user.ErrInvalidCredentials
	}

	u.ClearLockout()
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update user after login: %w", err)
	}

	accessToken, _, err := s.tokenSvc.GenerateAccessToken(ctx, u.ID, u.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	refreshToken, _, refreshExp, err := s.tokenSvc.GenerateRefreshToken(ctx, u.ID, u.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshExp,
		User:         u,
	}, nil
}

func (s *UserAuthService) VerifyEmail(ctx context.Context, tokenString string) error {
	claims, err := s.parseVerificationToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid verification token: %w", err)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return fmt.Errorf("invalid user id in token")
	}

	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if u.IsEmailVerified {
		return nil // already verified
	}

	u.VerifyEmail()
	return s.repo.Update(ctx, u)
}

type verificationClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

func (s *UserAuthService) generateVerificationToken(userID uuid.UUID, email string) (string, error) {
	claims := verificationClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.VerificationTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email: email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.VerificationTokenSecret))
}

func (s *UserAuthService) parseVerificationToken(tokenString string) (*verificationClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &verificationClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.VerificationTokenSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*verificationClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
