package repository

import (
	"context"

	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/google/uuid"
)

type AchievementRepository interface {
	GetAll(ctx context.Context) ([]*entity.Achievement, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.UserAchievement, error)
	CreateUserAchievement(ctx context.Context, userID, achievementID uuid.UUID) error
}
