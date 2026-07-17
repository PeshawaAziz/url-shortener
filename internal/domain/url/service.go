package url

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const maxGenerationRetries = 5

type ShortenerService struct {
	repo        URLRepository
	generator   SlugGenerator
	bloom       BloomChecker
	reserved    ReservedSlugChecker
	idempotency IdempotencyStore
}

func NewShortenerService(
	repo URLRepository,
	gen SlugGenerator,
	bloom BloomChecker,
	reserved ReservedSlugChecker,
	idempotency IdempotencyStore,
) *ShortenerService {
	return &ShortenerService{
		repo:        repo,
		generator:   gen,
		bloom:       bloom,
		reserved:    reserved,
		idempotency: idempotency,
	}
}

type ShortenURLInput struct {
	TenantID       uuid.UUID
	UserID         uuid.UUID
	OriginalURL    string
	DesiredSlug    string
	IdempotencyKey string
}

func (s *ShortenerService) ShortenURL(ctx context.Context, input ShortenURLInput) (*URL, error) {
	if input.IdempotencyKey != "" {
		existingID, err := s.idempotency.Get(ctx, input.IdempotencyKey)
		if err == nil {
			existingURL, err := s.repo.GetByID(ctx, existingID)
			if err == nil {
				return existingURL, nil
			}
		}
	}

	dest, err := NewOriginalURL(input.OriginalURL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	var slug Slug

	if input.DesiredSlug != "" {
		slug, err = NewSlug(ctx, input.DesiredSlug, s.reserved)
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

			deletedURL, err := s.repo.GetDeletedBySlug(ctx, input.TenantID, string(slug))
			if err == nil && deletedURL != nil {
				return nil, ErrSlugConflict
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
			// Race condition: someone grabbed the slug between our check and save.
			if input.DesiredSlug == "" {
				// If auto-gen, we can safely retry
				return s.ShortenURL(ctx, input)
			}
			return nil, ErrSlugConflict
		}
		return nil, err
	}

	s.bloom.Add(ctx, string(slug))

	if input.IdempotencyKey != "" {
		_ = s.idempotency.Save(ctx, input.IdempotencyKey, urlEntity.ID)
	}

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
