package model

import "time"

// Click описывает один переход по короткой ссылке.
type Click struct {
	ID        int64     `db:"id" json:"id"`
	LinkID    int64     `db:"link_id" json:"link_id"`
	ClickedAt time.Time `db:"clicked_at" json:"clicked_at"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
}
