package entity

import (
	"time"

	"github.com/google/uuid"
)

type Fragment struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Text      string
	Geomark   *Geomark
	SoundURL  string
	PhotoURLs []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Geomark struct {
	Lat float64
	Lng float64
}