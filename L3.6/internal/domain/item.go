package domain

import "time"

// Item описывает одну запись о доходе или расходе.
type Item struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Analytics хранит агрегированную аналитику по записям.
type Analytics struct {
	Sum    float64 `json:"sum"`
	Avg    float64 `json:"avg"`
	Count  int64   `json:"count"`
	Median float64 `json:"median"`
	P90    float64 `json:"p90"`
}

// ListFilter описывает фильтры и пагинацию списка записей.
type ListFilter struct {
	From     *time.Time
	To       *time.Time
	Type     string
	Category string
	SortBy   string
	Order    string
	Limit    int
	Offset   int
}

// AnalyticsFilter описывает фильтры для аналитики.
type AnalyticsFilter struct {
	From     *time.Time
	To       *time.Time
	Type     string
	Category string
}
