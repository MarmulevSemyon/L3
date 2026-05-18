package domain

import (
	"encoding/json"
	"time"
)

// ItemHistory описывает запись истории изменения товара.
type ItemHistory struct {
	ID                int64           `json:"id"`
	ItemID            int64           `json:"item_id"`
	Action            string          `json:"action"`
	ChangedByUserID   *int64          `json:"changed_by_user_id"`
	ChangedByUsername *string         `json:"changed_by_username"`
	ChangedByRole     *string         `json:"changed_by_role"`
	ChangedAt         time.Time       `json:"changed_at"`
	OldData           json.RawMessage `json:"old_data"`
	NewData           json.RawMessage `json:"new_data"`
}
