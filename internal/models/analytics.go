package models

type AnalyticsResponse struct {
	ShortURL    string           `json:"short_url"`
	TotalClicks int              `json:"total_clicks"`
	Clicks      []*Click         `json:"clicks,omitempty"`
	ByDay       []DayStats       `json:"by_day,omitempty"`
	ByMonth     []MonthStats     `json:"by_month,omitempty"`
	ByUserAgent []UserAgentStats `json:"by_user_agent,omitempty"`
}

type DayStats struct {
	Date   string `json:"date"`
	Clicks int    `json:"clicks"`
}

type MonthStats struct {
	Month  string `json:"month"`
	Clicks int    `json:"clicks"`
}

type UserAgentStats struct {
	UserAgent string `json:"user_agent"`
	Clicks    int    `json:"clicks"`
}
