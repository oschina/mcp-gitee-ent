package types

type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Public      int    `json:"public"`
	AvatarURL   string `json:"avatar_url"`
	Description string `json:"description"`
}
