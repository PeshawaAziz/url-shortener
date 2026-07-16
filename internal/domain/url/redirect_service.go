package url

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type RedirectService struct {
	repo     URLRepository
	cache    URLCache
	reporter ClickReporter
}

func NewRedirectService(repo URLRepository, cache URLCache, reporter ClickReporter) *RedirectService {
	return &RedirectService{
		repo:     repo,
		cache:    cache,
		reporter: reporter,
	}
}

func (s *RedirectService) Resolve(ctx context.Context, tenantID uuid.UUID, slug string, meta ClickMetadata) (*URL, error) {
	cachedURL, err := s.cache.Get(ctx, tenantID, slug)
	if err == nil && cachedURL != nil {
		if !cachedURL.CanRedirect(time.Now()) {
			return nil, ErrLinkExpired
		}
		s.reporter.RecordClick(ctx, cachedURL, meta)
		return cachedURL, nil
	}

	urlEntity, err := s.repo.GetBySlug(ctx, tenantID, slug)
	if errors.Is(err, ErrURLNotFound) {
		_ = s.cache.SetNegative(ctx, tenantID, slug)
		return nil, ErrURLNotFound
	}
	if err != nil {
		return nil, err
	}

	if !urlEntity.CanRedirect(time.Now()) {
		return nil, ErrLinkExpired
	}

	_ = s.cache.Set(ctx, urlEntity)

	s.reporter.RecordClick(ctx, urlEntity, meta)

	return urlEntity, nil
}
