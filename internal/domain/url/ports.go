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
}
