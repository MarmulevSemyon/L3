package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"commentTree/internal/model"
)

// PostgresRepository реализует работу с комментариями в PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgres создает новый репозиторий комментариев на PostgreSQL.
func NewPostgres(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create сохраняет новый комментарий в базе данных.
func (r *PostgresRepository) Create(ctx context.Context, comment *model.Comment) error {
	const query = `
		INSERT INTO comments (parent_id, author, body)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		comment.ParentID,
		comment.Author,
		comment.Body,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}

	return nil
}

// GetByID возвращает комментарий по его идентификатору.
func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*model.Comment, error) {
	const query = `
		SELECT id, parent_id, author, body, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	var comment model.Comment
	var parentID sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&parentID,
		&comment.Author,
		&comment.Body,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("get comment by id: %w", err)
	}

	if parentID.Valid {
		comment.ParentID = &parentID.Int64
	}

	return &comment, nil
}

// GetChildren возвращает комментарии одного уровня вложенности по parentID.
func (r *PostgresRepository) GetChildren(ctx context.Context, parentID *int64, limit, offset int, sortBy, order string) ([]model.Comment, int, error) {
	sortBy = normalizeSortBy(sortBy)
	order = normalizeOrder(order)

	countQuery, dataQuery := buildChildrenQueries(sortBy, order)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, parentID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count children: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, dataQuery, parentID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get children: %w", err)
	}
	defer rows.Close()

	comments, err := scanComments(rows)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// GetAllRoots возвращает корневые комментарии.
func (r *PostgresRepository) GetAllRoots(ctx context.Context, limit, offset int, sortBy, order string) ([]model.Comment, int, error) {
	sortBy = normalizeSortBy(sortBy)
	order = normalizeOrder(order)

	countQuery, dataQuery := buildRootsQueries(sortBy, order)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count roots: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, dataQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get roots: %w", err)
	}
	defer rows.Close()

	comments, err := scanComments(rows)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// DeleteSubtree удаляет комментарий и всех его потомков.
func (r *PostgresRepository) DeleteSubtree(ctx context.Context, id int64) error {
	const query = `
		WITH RECURSIVE subtree AS (
			SELECT id
			FROM comments
			WHERE id = $1

			UNION ALL

			SELECT c.id
			FROM comments c
			INNER JOIN subtree s ON c.parent_id = s.id
		)
		DELETE FROM comments
		WHERE id IN (SELECT id FROM subtree)
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete subtree: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete subtree rows affected: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Search выполняет полнотекстовый поиск по комментариям.
func (r *PostgresRepository) Search(ctx context.Context, queryText string, limit, offset int, sortBy, order string) ([]model.Comment, int, error) {
	sortBy = normalizeSortBy(sortBy)
	order = normalizeOrder(order)

	countQuery, dataQuery := buildSearchQueries(sortBy, order)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, queryText).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count search results: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, dataQuery, queryText, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("search comments: %w", err)
	}
	defer rows.Close()

	comments, err := scanComments(rows)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func buildChildrenQueries(sortBy, order string) (string, string) {
	countQuery := `
		SELECT COUNT(*)
		FROM comments
		WHERE parent_id = $1
	`

	dataQuery := fmt.Sprintf(`
		SELECT id, parent_id, author, body, created_at, updated_at
		FROM comments
		WHERE parent_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortBy, order)

	return countQuery, dataQuery
}

func buildRootsQueries(sortBy, order string) (string, string) {
	countQuery := `
		SELECT COUNT(*)
		FROM comments
		WHERE parent_id IS NULL
	`

	dataQuery := fmt.Sprintf(`
		SELECT id, parent_id, author, body, created_at, updated_at
		FROM comments
		WHERE parent_id IS NULL
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, sortBy, order)

	return countQuery, dataQuery
}

func buildSearchQueries(sortBy, order string) (string, string) {
	countQuery := `
		SELECT COUNT(*)
		FROM comments
		WHERE search_vector @@ plainto_tsquery('simple', $1)
	`

	dataQuery := fmt.Sprintf(`
		SELECT id, parent_id, author, body, created_at, updated_at
		FROM comments
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortBy, order)

	return countQuery, dataQuery
}

func scanComments(rows *sql.Rows) ([]model.Comment, error) {
	var comments []model.Comment

	for rows.Next() {
		var comment model.Comment
		var parentID sql.NullInt64

		err := rows.Scan(
			&comment.ID,
			&parentID,
			&comment.Author,
			&comment.Body,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}

		if parentID.Valid {
			comment.ParentID = &parentID.Int64
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return comments, nil
}

func normalizeSortBy(sortBy string) string {
	if s := strings.ToLower(sortBy); s == "updated_at" {
		return s
	}
	return "created_at"
}

func normalizeOrder(order string) string {
	if strings.ToLower(order) == "asc" {
		return "ASC"
	}
	return "DESC"
}
