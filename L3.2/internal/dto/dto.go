package dto

import "time"

// ShortenRequest описывает запрос на создание короткой ссылки.
type ShortenRequest struct {
	URL       string     `json:"url"`
	Alias     string     `json:"alias,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ShortenResponse описывает ответ после создания короткой ссылки.
type ShortenResponse struct {
	Alias string `json:"alias"`
	Short string `json:"short_url"`
}

// RawClickResponse описывает одну запись о переходе по короткой ссылке.
type RawClickResponse struct {
	ClickedAt time.Time `json:"clicked_at"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
}

// DayAnalyticsItem описывает агрегированную статистику переходов по дням.
type DayAnalyticsItem struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

// MonthAnalyticsItem описывает агрегированную статистику переходов по месяцам.
type MonthAnalyticsItem struct {
	Month  string `json:"month"`
	Clicks int64  `json:"clicks"`
}

// UserAgentAnalyticsItem описывает агрегированную статистику переходов по User-Agent.
type UserAgentAnalyticsItem struct {
	UserAgent string `json:"user_agent"`
	Clicks    int64  `json:"clicks"`
}
