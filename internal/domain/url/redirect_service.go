package url

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type RedirectService struct {
	repo        URLRepository
	cache       URLCache
	reporter    ClickReporter
	rulesEngine *RulesEngine
	rateLimiter RateLimiter
}

func NewRedirectService(
	repo URLRepository,
	cache URLCache,
	reporter ClickReporter,
	rulesEngine *RulesEngine,
	rateLimiter RateLimiter,
) *RedirectService {
	return &RedirectService{
		repo:        repo,
		cache:       cache,
		reporter:    reporter,
		rulesEngine: rulesEngine,
		rateLimiter: rateLimiter,
	}
}

func (s *RedirectService) Resolve(ctx context.Context, tenantID uuid.UUID, slug string, reqCtx RequestContext) (*URL, string, error) {
	cachedURL, err := s.cache.Get(ctx, tenantID, slug)
	if err == nil && cachedURL != nil {
		return s.processRedirect(ctx, cachedURL, reqCtx)
	}

	urlEntity, err := s.repo.GetBySlug(ctx, tenantID, slug)
	if errors.Is(err, ErrURLNotFound) {
		_ = s.cache.SetNegative(ctx, tenantID, slug)
		return nil, "", ErrURLNotFound
	}
	if err != nil {
		return nil, "", err
	}

	_ = s.cache.Set(ctx, urlEntity)
	return s.processRedirect(ctx, urlEntity, reqCtx)
}

func (s *RedirectService) processRedirect(ctx context.Context, u *URL, reqCtx RequestContext) (*URL, string, error) {
	if !u.CanRedirect(time.Now()) {
		return nil, "", ErrLinkExpired
	}

	if u.HasRateLimit() {
		allowed, err := s.rateLimiter.Allow(ctx, u.TenantID, string(u.Slug), *u.RateLimitPerHr)
		if err != nil {
			// Fail open on rate limiter error to not break redirects, but log it
		} else if !allowed {
			return nil, "", ErrRateLimited
		}
	}

	if u.IsPasswordProtected() {
		return u, "", ErrPasswordRequired
	}

	finalDestination := s.rulesEngine.ResolveDestination(u, reqCtx)

	meta := ClickMetadata{
		IPAddress: reqCtx.IPAddress,
		UserAgent: reqCtx.UserAgent,
	}
	s.reporter.RecordClick(ctx, u, meta)

	return u, finalDestination, nil
}
