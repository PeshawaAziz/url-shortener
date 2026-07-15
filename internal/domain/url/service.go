package url

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const maxGenerationRetries = 5

type ShortenerService struct {
	repo      URLRepository
	generator SlugGenerator
	bloom     BloomChecker
}

func NewShortenerService(repo URLRepository, gen SlugGenerator, bloom BloomChecker) *ShortenerService {
	return &ShortenerService{
		repo:      repo,
		generator: gen,
		bloom:     bloom,
	}
}

type ShortenURLInput struct {
	TenantID    uuid.UUID
	UserID      uuid.UUID
	OriginalURL string
	DesiredSlug string // Optional. If empty, auto-generate.
}

func (s *ShortenerService) ShortenURL(ctx context.Context, input ShortenURLInput) (*URL, error) {
	dest, err := NewOriginalURL(input.OriginalURL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	var slug Slug

	if input.DesiredSlug != "" {
		slug, err = NewSlug(input.DesiredSlug)
		if err != nil {
			return nil, ErrInvalidSlug
		}

		if s.bloom.Exists(ctx, string(slug)) {
			_, err := s.repo.GetBySlug(ctx, input.TenantID, string(slug))
			if err == nil {
				return nil, ErrSlugConflict
			}
			if !errors.Is(err, ErrURLNotFound) {
				return nil, err
			}
		}
	} else {
		var err error
		slug, err = s.generateUniqueSlug(ctx, input.TenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate unique slug: %w", err)
		}
	}

	urlEntity := NewURL(input.TenantID, input.UserID, slug, dest)

	err = s.repo.Save(ctx, urlEntity)
	if err != nil {
		if errors.Is(err, ErrSlugConflict) {
			if input.DesiredSlug == "" {
				// If auto-gen, we can safely retry
				return s.ShortenURL(ctx, input)
			}
			return nil, ErrSlugConflict
		}
		return nil, err
	}

	s.bloom.Add(ctx, string(slug))

	return urlEntity, nil
}

func (s *ShortenerService) generateUniqueSlug(ctx context.Context, tenantID uuid.UUID) (Slug, error) {
	for i := 0; i < maxGenerationRetries; i++ {
		rawSlug := s.generator.Generate()

		if s.bloom.Exists(ctx, rawSlug) {
			continue
		}

		_, err := s.repo.GetBySlug(ctx, tenantID, rawSlug)
		if errors.Is(err, ErrURLNotFound) {
			return Slug(rawSlug), nil
		}
		if err != nil {
			return "", fmt.Errorf("db error during slug check: %w", err)
		}
	}

	return "", ErrMaxRetriesHit
}
