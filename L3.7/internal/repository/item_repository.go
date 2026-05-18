package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"warehousecontrol/internal/domain"

	"github.com/wb-go/wbf/dbpg"
)

// ItemRepository работает с товарами в PostgreSQL.
type ItemRepository struct {
	db *dbpg.DB
}

// NewItemRepository создаёт репозиторий товаров.
func NewItemRepository(db *dbpg.DB) *ItemRepository {
	return &ItemRepository{
		db: db,
	}
}

// Create создаёт новый товар.
func (r *ItemRepository) Create(
	ctx context.Context,
	actor domain.Actor,
	input domain.CreateItemInput,
) (domain.Item, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Item{}, fmt.Errorf("begin tx: %w", err)
	}
	defer rollbackTx(tx)

	if err := setAuditContext(ctx, tx, actor); err != nil {
		return domain.Item{}, fmt.Errorf("set audit context: %w", err)
	}

	const query = `
		INSERT INTO items (name, description, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, quantity, price, created_at, updated_at;
	`

	var item domain.Item

	err = tx.QueryRowContext(
		ctx,
		query,
		input.Name,
		input.Description,
		input.Quantity,
		input.Price,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Quantity,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Item{}, fmt.Errorf("create item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.Item{}, fmt.Errorf("commit tx: %w", err)
	}

	return item, nil
}

// List возвращает список всех товаров.
func (r *ItemRepository) List(ctx context.Context) ([]domain.Item, error) {
	const query = `
		SELECT id, name, description, quantity, price, created_at, updated_at
		FROM items
		ORDER BY id;
	`

	rows, err := r.db.Master.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []domain.Item

	for rows.Next() {
		var item domain.Item

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Quantity,
			&item.Price,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}

// GetByID возвращает товар по ID.
func (r *ItemRepository) GetByID(ctx context.Context, id int64) (domain.Item, error) {
	const query = `
		SELECT id, name, description, quantity, price, created_at, updated_at
		FROM items
		WHERE id = $1;
	`

	var item domain.Item

	err := r.db.Master.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Quantity,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Item{}, ErrNotFound
		}

		return domain.Item{}, fmt.Errorf("get item by id: %w", err)
	}

	return item, nil
}

// Update обновляет товар.
func (r *ItemRepository) Update(
	ctx context.Context,
	actor domain.Actor,
	id int64,
	input domain.UpdateItemInput,
) (domain.Item, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Item{}, fmt.Errorf("begin tx: %w", err)
	}
	defer rollbackTx(tx)

	if err := setAuditContext(ctx, tx, actor); err != nil {
		return domain.Item{}, fmt.Errorf("set audit context: %w", err)
	}

	const query = `
		UPDATE items
		SET name = $1,
		    description = $2,
		    quantity = $3,
		    price = $4
		WHERE id = $5
		RETURNING id, name, description, quantity, price, created_at, updated_at;
	`

	var item domain.Item

	err = tx.QueryRowContext(
		ctx,
		query,
		input.Name,
		input.Description,
		input.Quantity,
		input.Price,
		id,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Quantity,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Item{}, ErrNotFound
		}

		return domain.Item{}, fmt.Errorf("update item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.Item{}, fmt.Errorf("commit tx: %w", err)
	}

	return item, nil
}

// Delete удаляет товар.
func (r *ItemRepository) Delete(
	ctx context.Context,
	actor domain.Actor,
	id int64,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer rollbackTx(tx)

	if err := setAuditContext(ctx, tx, actor); err != nil {
		return fmt.Errorf("set audit context: %w", err)
	}

	const query = `
		DELETE FROM items
		WHERE id = $1;
	`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func setAuditContext(ctx context.Context, tx *sql.Tx, actor domain.Actor) error {
	const query = `
		SELECT
			set_config('app.user_id', $1, true),
			set_config('app.username', $2, true),
			set_config('app.role', $3, true);
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		fmt.Sprintf("%d", actor.ID),
		actor.Username,
		actor.Role,
	)

	return err
}

func rollbackTx(tx *sql.Tx) {
	if tx == nil {
		return
	}

	_ = tx.Rollback()
}
