package repository

import (
	"context"

	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/google/uuid"
)

type FragmentRepository interface {
	Create(ctx context.Context, fragment *entity.Fragment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Fragment, error)
	List(ctx context.Context, userID uuid.UUID, lat, lng, radius float64) ([]*entity.Fragment, error)
	Update(ctx context.Context, fragment *entity.Fragment) error
	Delete(ctx context.Context, id uuid.UUID) error
}