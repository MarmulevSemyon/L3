package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"salestracker/internal/domain"

	"github.com/wb-go/wbf/dbpg"
)

// ItemRepository работает с таблицей items в PostgreSQL.
type ItemRepository struct {
	db *dbpg.DB
}

// NewItemRepository создаёт репозиторий записей.
func NewItemRepository(db *dbpg.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// Create создаёт новую запись.
func (r *ItemRepository) Create(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	const query = `
		INSERT INTO items (type, amount, category, description, occurred_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, type, amount, category, description, occurred_at, created_at, updated_at
	`

	var out domain.Item

	err := r.db.Master.QueryRowContext(
		ctx,
		query,
		item.Type,
		item.Amount,
		item.Category,
		item.Description,
		item.OccurredAt,
	).Scan(
		&out.ID,
		&out.Type,
		&out.Amount,
		&out.Category,
		&out.Description,
		&out.OccurredAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// GetByID возвращает запись по ID.
func (r *ItemRepository) GetByID(ctx context.Context, id int64) (*domain.Item, error) {
	const query = `
		SELECT id, type, amount, category, description, occurred_at, created_at, updated_at
		FROM items
		WHERE id = $1
	`

	var item domain.Item

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Type,
		&item.Amount,
		&item.Category,
		&item.Description,
		&item.OccurredAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// List возвращает список записей с фильтрами.
func (r *ItemRepository) List(ctx context.Context, filter domain.ListFilter) ([]domain.Item, error) {
	sortBy := normalizeSortBy(filter.SortBy)
	order := normalizeOrder(filter.Order)

	query := fmt.Sprintf(`
		SELECT id, type, amount, category, description, occurred_at, created_at, updated_at
		FROM items
		WHERE ($1::timestamptz IS NULL OR occurred_at >= $1)
		  AND ($2::timestamptz IS NULL OR occurred_at <= $2)
		  AND ($3::text = '' OR type = $3)
		  AND ($4::text = '' OR category = $4)
		ORDER BY %s %s
		LIMIT $5 OFFSET $6
	`, sortBy, order)

	rows, err := r.db.QueryContext(
		ctx,
		query,
		timeArg(filter.From),
		timeArg(filter.To),
		filter.Type,
		filter.Category,
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Item, 0)

	for rows.Next() {
		var item domain.Item

		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Amount,
			&item.Category,
			&item.Description,
			&item.OccurredAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// Update обновляет запись по ID.
func (r *ItemRepository) Update(ctx context.Context, id int64, item *domain.Item) (*domain.Item, error) {
	const query = `
		UPDATE items
		SET type = $1,
		    amount = $2,
		    category = $3,
		    description = $4,
		    occurred_at = $5,
		    updated_at = NOW()
		WHERE id = $6
		RETURNING id, type, amount, category, description, occurred_at, created_at, updated_at
	`

	var out domain.Item

	err := r.db.Master.QueryRowContext(
		ctx,
		query,
		item.Type,
		item.Amount,
		item.Category,
		item.Description,
		item.OccurredAt,
		id,
	).Scan(
		&out.ID,
		&out.Type,
		&out.Amount,
		&out.Category,
		&out.Description,
		&out.OccurredAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// Delete удаляет запись по ID.
func (r *ItemRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM items WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Analytics возвращает агрегированную аналитику.
func (r *ItemRepository) Analytics(ctx context.Context, filter domain.AnalyticsFilter) (*domain.Analytics, error) {
	const query = `
		SELECT
			COALESCE(SUM(amount), 0) AS sum,
			COALESCE(AVG(amount), 0) AS avg,
			COUNT(*) AS count,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount), 0) AS median,
			COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount), 0) AS p90
		FROM items
		WHERE ($1::timestamptz IS NULL OR occurred_at >= $1)
		  AND ($2::timestamptz IS NULL OR occurred_at <= $2)
		  AND ($3::text = '' OR type = $3)
		  AND ($4::text = '' OR category = $4)
	`

	var out domain.Analytics

	err := r.db.QueryRowContext(
		ctx,
		query,
		timeArg(filter.From),
		timeArg(filter.To),
		filter.Type,
		filter.Category,
	).Scan(
		&out.Sum,
		&out.Avg,
		&out.Count,
		&out.Median,
		&out.P90,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func normalizeSortBy(sortBy string) string {
	switch sortBy {
	case "amount":
		return "amount"
	case "created_at":
		return "created_at"
	default:
		return "occurred_at"
	}
}

func normalizeOrder(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}

	return "DESC"
}

func timeArg(t *time.Time) any {
	if t == nil {
		return nil
	}

	return *t
}
