package url

import (
	"context"

	"github.com/google/uuid"
)

type SlugGenerator interface {
	Generate() string
}

type BloomChecker interface {
	Exists(ctx context.Context, slug string) bool
	Add(ctx context.Context, slug string)
}

type URLRepository interface {
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*URL, error)
	Save(ctx context.Context, u *URL) error
	GetByID(ctx context.Context, id uuid.UUID) (*URL, error)
	Update(ctx context.Context, u *URL) error
	GetDeletedBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*URL, error)
	ListActiveForSweep(ctx context.Context, limit, offset int) ([]*URL, error)
}

type ReservedSlugChecker interface {
	IsReserved(ctx context.Context, slug string) bool
}

type IdempotencyStore interface {
	Get(ctx context.Context, key string) (uuid.UUID, error)
	Save(ctx context.Context, key string, urlID uuid.UUID) error
}

type URLCache interface {
	Get(ctx context.Context, tenantID uuid.UUID, slug string) (*URL, error)
	Set(ctx context.Context, u *URL) error
	SetNegative(ctx context.Context, tenantID uuid.UUID, slug string) error
}

type ClickReporter interface {
	RecordClick(ctx context.Context, u *URL, metadata ClickMetadata)
}

type ClickMetadata struct {
	IPAddress string
	UserAgent string
	Referer   string
}
