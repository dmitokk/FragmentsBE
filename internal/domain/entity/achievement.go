package entity

import (
	"time"

	"github.com/google/uuid"
)

type Achievement struct {
	ID             uuid.UUID
	Code           string
	Name           string
	Description    string
	IconURL        string
	ConditionType  string
	ConditionValue int
	CreatedAt      time.Time
}

type UserAchievement struct {
	UserID        uuid.UUID
	AchievementID uuid.UUID
	UnlockedAt    time.Time
}
