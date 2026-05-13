package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/google/uuid"
)

type AchievementRepository struct {
	db *sql.DB
}

func NewAchievementRepository(db *sql.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func (r *AchievementRepository) GetAll(ctx context.Context) ([]*entity.Achievement, error) {
	query := `
		SELECT id, code, name, description, icon_url, condition_type, condition_value, created_at
		FROM achievements
		ORDER BY condition_value ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list achievements: %w", err)
	}
	defer rows.Close()

	var achievements []*entity.Achievement
	for rows.Next() {
		var a entity.Achievement
		var iconURL sql.NullString

		err := rows.Scan(
			&a.ID, &a.Code, &a.Name, &a.Description,
			&iconURL, &a.ConditionType, &a.ConditionValue, &a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}

		if iconURL.Valid {
			a.IconURL = iconURL.String
		}

		achievements = append(achievements, &a)
	}

	return achievements, nil
}

func (r *AchievementRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.UserAchievement, error) {
	query := `
		SELECT user_id, achievement_id, unlocked_at
		FROM user_achievements
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user achievements: %w", err)
	}
	defer rows.Close()

	var achievements []*entity.UserAchievement
	for rows.Next() {
		var ua entity.UserAchievement
		err := rows.Scan(&ua.UserID, &ua.AchievementID, &ua.UnlockedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user achievement: %w", err)
		}
		achievements = append(achievements, &ua)
	}

	return achievements, nil
}

func (r *AchievementRepository) CreateUserAchievement(ctx context.Context, userID, achievementID uuid.UUID) error {
	query := `
		INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, achievementID)
	if err != nil {
		return fmt.Errorf("failed to create user achievement: %w", err)
	}

	return nil
}
