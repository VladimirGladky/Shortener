package models

type Click struct {
	ID        int    `json:"id"`
	ShortCode string `json:"short_code"`
	ClickedAt string `json:"clicked_at"`
	UserAgent string `json:"user_agent"`
	IpAddress string `json:"ip_address"`
}
