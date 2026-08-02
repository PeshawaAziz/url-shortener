package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
}
