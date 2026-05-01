package repository

import (
	"context"

	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}