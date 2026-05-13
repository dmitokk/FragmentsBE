package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/dmitokk/FragmentsBE/internal/application/dto"
	"github.com/dmitokk/FragmentsBE/internal/domain/repository"
	"github.com/google/uuid"
)

type AchievementService struct {
	achievementRepo  repository.AchievementRepository
	userFragmentRepo repository.UserFragmentRepository
	fragmentRepo     repository.FragmentRepository
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	userFragmentRepo repository.UserFragmentRepository,
	fragmentRepo repository.FragmentRepository,
) *AchievementService {
	return &AchievementService{
		achievementRepo:  achievementRepo,
		userFragmentRepo: userFragmentRepo,
		fragmentRepo:     fragmentRepo,
	}
}

func (s *AchievementService) GetAll(ctx context.Context, userID uuid.UUID) ([]*dto.AchievementResponse, error) {
	achievements, err := s.achievementRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	userAchievements, err := s.achievementRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	unlockedMap := make(map[uuid.UUID]time.Time)
	for _, ua := range userAchievements {
		unlockedMap[ua.AchievementID] = ua.UnlockedAt
	}

	responses := make([]*dto.AchievementResponse, len(achievements))
	for i, a := range achievements {
		unlockedAt, completed := unlockedMap[a.ID]
		resp := &dto.AchievementResponse{
			ID:          a.ID.String(),
			Code:        a.Code,
			Name:        a.Name,
			Description: a.Description,
			IconURL:     a.IconURL,
			IsCompleted: completed,
		}
		if completed {
			resp.UnlockedAt = unlockedAt.Format(time.RFC3339)
		}
		responses[i] = resp
	}

	return responses, nil
}

func (s *AchievementService) GetMine(ctx context.Context, userID uuid.UUID) ([]*dto.AchievementResponse, error) {
	all, err := s.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	var mine []*dto.AchievementResponse
	for _, a := range all {
		if a.IsCompleted {
			mine = append(mine, a)
		}
	}

	return mine, nil
}

func (s *AchievementService) CheckAndUnlock(ctx context.Context, userID uuid.UUID, fragmentID uuid.UUID) error {
	allAchievements, err := s.achievementRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	userAchievements, err := s.achievementRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	unlockedMap := make(map[uuid.UUID]bool)
	for _, ua := range userAchievements {
		unlockedMap[ua.AchievementID] = true
	}

	fragment, err := s.fragmentRepo.GetByID(ctx, fragmentID)
	if err != nil {
		return err
	}

	userFragments, err := s.userFragmentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, achievement := range allAchievements {
		if unlockedMap[achievement.ID] {
			continue
		}

		unlocked := false

		switch achievement.ConditionType {
		case "found_count":
			unlocked = len(userFragments) >= achievement.ConditionValue

		case "has_photo":
			unlocked = len(fragment.PhotoURLs) > 0

		case "has_sound":
			unlocked = fragment.SoundURL != ""
		}

		if unlocked {
			err := s.achievementRepo.CreateUserAchievement(ctx, userID, achievement.ID)
			if err != nil {
				slog.Error("Failed to unlock achievement", "achievement", achievement.Code, "error", err)
				continue
			}
			slog.Info("Achievement unlocked", "user", userID, "achievement", achievement.Code)
		}
	}

	return nil
}
