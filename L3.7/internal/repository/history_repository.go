package repository

import (
	"context"
	"database/sql"
	"fmt"

	"warehousecontrol/internal/domain"

	"github.com/wb-go/wbf/dbpg"
)

// HistoryRepository работает с историей изменений товаров.
type HistoryRepository struct {
	db *dbpg.DB
}

// NewHistoryRepository создаёт репозиторий истории изменений.
func NewHistoryRepository(db *dbpg.DB) *HistoryRepository {
	return &HistoryRepository{
		db: db,
	}
}

// ListByItemID возвращает историю изменений конкретного товара.
func (r *HistoryRepository) ListByItemID(ctx context.Context, itemID int64) ([]domain.ItemHistory, error) {
	const query = `
		SELECT
		    id,
		    item_id,
		    action,
		    changed_by_user_id,
		    changed_by_username,
		    changed_by_role,
		    changed_at,
		    old_data,
		    new_data
		FROM item_history
		WHERE item_id = $1
		ORDER BY id DESC;
	`

	rows, err := r.db.Master.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("list item history: %w", err)
	}
	defer rows.Close()

	var history []domain.ItemHistory

	for rows.Next() {
		var record domain.ItemHistory

		var changedByUserID sql.NullInt64
		var changedByUsername sql.NullString
		var changedByRole sql.NullString
		var oldData []byte
		var newData []byte

		if err := rows.Scan(
			&record.ID,
			&record.ItemID,
			&record.Action,
			&changedByUserID,
			&changedByUsername,
			&changedByRole,
			&record.ChangedAt,
			&oldData,
			&newData,
		); err != nil {
			return nil, fmt.Errorf("scan item history: %w", err)
		}

		if changedByUserID.Valid {
			record.ChangedByUserID = &changedByUserID.Int64
		}

		if changedByUsername.Valid {
			record.ChangedByUsername = &changedByUsername.String
		}

		if changedByRole.Valid {
			record.ChangedByRole = &changedByRole.String
		}

		record.OldData = oldData
		record.NewData = newData

		history = append(history, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return history, nil
}
