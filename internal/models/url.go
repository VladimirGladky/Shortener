package models

import "time"

type Url struct {
	ID          int       `json:"id"`
	ShortUrl    string    `json:"short_url"`
	OriginalUrl string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}
