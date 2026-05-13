package repository

import (
	"context"

	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/google/uuid"
)

type UserFragmentRepository interface {
	Create(ctx context.Context, userID, fragmentID uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.UserFragment, error)
	Exists(ctx context.Context, userID, fragmentID uuid.UUID) (bool, error)
}
