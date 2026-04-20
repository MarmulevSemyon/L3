package dto

import "time"

// CreateItemRequest используется для создания записи.
type CreateItemRequest struct {
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// UpdateItemRequest используется для обновления записи.
type UpdateItemRequest struct {
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
}
