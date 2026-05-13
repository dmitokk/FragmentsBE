package dto

type AchievementResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url,omitempty"`
	IsCompleted bool   `json:"is_completed"`
	UnlockedAt  string `json:"unlocked_at,omitempty"`
}
