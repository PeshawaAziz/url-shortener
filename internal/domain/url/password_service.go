package url

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type PasswordService struct {
	repo   URLRepository
	hasher PasswordHasher
}

func NewPasswordService(repo URLRepository, hasher PasswordHasher) *PasswordService {
	return &PasswordService{repo: repo, hasher: hasher}
}

func (s *PasswordService) VerifyPassword(ctx context.Context, tenantID uuid.UUID, slug, password string) (bool, error) {
	u, err := s.repo.GetBySlug(ctx, tenantID, slug)
	if err != nil {
		return false, err
	}

	if !u.IsPasswordProtected() {
		return true, nil
	}

	if !s.hasher.Compare(u.PasswordHash, password) {
		return false, errors.New("invalid password")
	}

	return true, nil
}
