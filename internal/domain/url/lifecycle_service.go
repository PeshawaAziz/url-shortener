package url

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LifecycleService struct {
	repo URLRepository
}

func NewLifecycleService(repo URLRepository) *LifecycleService {
	return &LifecycleService{repo: repo}
}

func (s *LifecycleService) PauseLink(ctx context.Context, id uuid.UUID) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.Pause(); err != nil {
		return err
	}
	return s.repo.Update(ctx, u)
}

func (s *LifecycleService) ResumeLink(ctx context.Context, id uuid.UUID) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.Resume(); err != nil {
		return err
	}
	return s.repo.Update(ctx, u)
}

func (s *LifecycleService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.SoftDelete(); err != nil {
		return err
	}
	return s.repo.Update(ctx, u)
}

func (s *LifecycleService) SweepExpired(ctx context.Context, batchSize int) (int, error) {
	urlsToExpire, err := s.repo.ListActiveForSweep(ctx, batchSize, 0)
	if err != nil {
		return 0, err
	}

	expiredCount := 0
	for _, u := range urlsToExpire {
		if !u.CanRedirect(time.Now()) {
			u.State = StateExpired
			u.UpdatedAt = time.Now()
			if err := s.repo.Update(ctx, u); err != nil {
				fmt.Printf("Failed to expire URL %s: %v\n", u.ID, err)
				continue
			}
			expiredCount++
		}
	}
	return expiredCount, nil
}
