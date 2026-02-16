package models

type Url struct {
	ID          int    `json:"id"`
	ShortCode   string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
}
