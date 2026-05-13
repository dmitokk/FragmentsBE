package dto

import "github.com/google/uuid"

const DefaultLifetimeHours = 720

type CreateFragmentRequest struct {
	Text           string   `form:"text" json:"text" binding:"required"`
	Lat            *float64 `form:"lat" json:"lat" binding:"required"`
	Lng            *float64 `form:"lng" json:"lng" binding:"required"`
	ExpiresInHours *int     `form:"expires_in_hours" json:"expires_in_hours"`
}

func (r *CreateFragmentRequest) GetLifetimeHours() int {
	if r.ExpiresInHours != nil && *r.ExpiresInHours > 0 {
		return *r.ExpiresInHours
	}
	return DefaultLifetimeHours
}

type FragmentResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Text      string    `json:"text"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	SoundURL  string    `json:"sound_url,omitempty"`
	PhotoURLs []string  `json:"photo_urls,omitempty"`
	ExpiresAt string    `json:"expires_at"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}