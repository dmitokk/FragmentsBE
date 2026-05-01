package dto

import "github.com/google/uuid"

type CreateFragmentRequest struct {
	Text string  `form:"text" json:"text" binding:"required"`
	Lat  *float64 `form:"lat" json:"lat" binding:"required"`
	Lng  *float64 `form:"lng" json:"lng" binding:"required"`
}

type UpdateFragmentRequest struct {
	Text string  `json:"text"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type FragmentResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Text      string     `json:"text"`
	Lat       float64    `json:"lat"`
	Lng       float64    `json:"lng"`
	SoundURL  string     `json:"sound_url,omitempty"`
	PhotoURLs []string   `json:"photo_urls,omitempty"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}