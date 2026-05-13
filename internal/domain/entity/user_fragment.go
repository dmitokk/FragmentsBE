package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserFragment struct {
	UserID     uuid.UUID
	FragmentID uuid.UUID
	FoundAt    time.Time
}
