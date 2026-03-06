package models

import "time"

type Click struct {
	ID        int       `json:"id"`
	ShortUrl  string    `json:"short_url"`
	ClickedAt time.Time `json:"clicked_at"`
	UserAgent string    `json:"user_agent"`
}
