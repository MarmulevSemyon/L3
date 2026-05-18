package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"warehousecontrol/internal/domain"

	"github.com/wb-go/wbf/dbpg"
)

// UserRepository работает с пользователями.
type UserRepository struct {
	db *dbpg.DB
}

// NewUserRepository создаёт репозиторий пользователей.
func NewUserRepository(db *dbpg.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetActorByUsername возвращает пользователя и его роль по username.
func (r *UserRepository) GetActorByUsername(ctx context.Context, username string) (domain.Actor, error) {
	const query = `
		SELECT
		    users.id,
		    users.username,
		    roles.name
		FROM users
		JOIN roles ON roles.id = users.role_id
		WHERE users.username = $1;
	`

	var actor domain.Actor

	err := r.db.Master.QueryRowContext(ctx, query, username).Scan(
		&actor.ID,
		&actor.Username,
		&actor.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Actor{}, ErrNotFound
		}

		return domain.Actor{}, fmt.Errorf("get actor by username: %w", err)
	}

	return actor, nil
}
