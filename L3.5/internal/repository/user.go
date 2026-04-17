package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"

	"eventBooker/internal/domain"
)

// UserRepository предоставляет методы работы с пользователями в базе данных.
type UserRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

// NewUserRepo создаёт репозиторий для работы с пользователями.
func NewUserRepo(db *dbpg.DB) *UserRepository {
	return &UserRepository{
		db: db,
		strategy: retry.Strategy{
			Attempts: 3,
			Delay:    500 * time.Millisecond,
			Backoff:  2,
		},
	}
}

// Create сохраняет нового пользователя в базе данных.
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, username, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecWithRetry(ctx, r.strategy, query, u.ID, u.Username, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// GetByID возвращает пользователя по его идентификатору.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, username, created_at, updated_at FROM users WHERE id = $1`
	row, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, id)
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	var u domain.User
	if err = row.Scan(&u.ID, &u.Username, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &u, nil
}

// List возвращает список всех пользователей.
func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, username, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var res []*domain.User
	for rows.Next() {
		var u domain.User
		if err = rows.Scan(&u.ID, &u.Username, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		res = append(res, &u)
	}

	return res, rows.Err()
}
