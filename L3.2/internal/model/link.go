package model

import "time"

// Link описывает сокращённую ссылку и её исходный адрес.
type Link struct {
	ID          int64      `db:"id" json:"id"`
	Alias       string     `db:"alias" json:"alias"`
	OriginalURL string     `db:"original_url" json:"original_url"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
}
