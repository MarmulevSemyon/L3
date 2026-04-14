package repository

import (
	"context"
	"database/sql"
	"errors"
	"imageProcessor/internal/model"
)

// ErrImageNotFound возвращается, когда изображение не найдено в базе данных.
var ErrImageNotFound = errors.New("image not found")

// PostgresRepository реализует Repository через PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository создает новый репозиторий.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateImage сохраняет запись об изображении.
func (r *PostgresRepository) CreateImage(ctx context.Context, image *model.Image) error {
	const query = `
		INSERT INTO images (
			id, original_name, content_type, status,
			operation_resize, operation_thumb, operation_watermark,
			original_path, processed_path, thumb_path, error_text,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10, $11,
			NOW(), NOW()
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		image.ID,
		image.OriginalName,
		image.ContentType,
		image.Status,
		image.OperationResize,
		image.OperationThumb,
		image.OperationWatermark,
		image.OriginalPath,
		image.ProcessedPath,
		image.ThumbPath,
		image.ErrorText,
	)

	return err
}

// GetImageByID возвращает изображение по id.
func (r *PostgresRepository) GetImageByID(ctx context.Context, id string) (*model.Image, error) {
	const query = `
		SELECT
			id, original_name, content_type, status,
			operation_resize, operation_thumb, operation_watermark,
			original_path, processed_path, thumb_path, error_text,
			created_at, updated_at
		FROM images
		WHERE id = $1
	`

	var image model.Image

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&image.ID,
		&image.OriginalName,
		&image.ContentType,
		&image.Status,
		&image.OperationResize,
		&image.OperationThumb,
		&image.OperationWatermark,
		&image.OriginalPath,
		&image.ProcessedPath,
		&image.ThumbPath,
		&image.ErrorText,
		&image.CreatedAt,
		&image.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}

	return &image, nil
}

// UpdateImageStatus обновляет статус и текст ошибки.
func (r *PostgresRepository) UpdateImageStatus(ctx context.Context, id string, status string, errText string) error {
	const query = `
		UPDATE images
		SET status = $2, error_text = $3, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, status, errText)
	return err
}

// UpdateImageResult обновляет результат обработки.
func (r *PostgresRepository) UpdateImageResult(ctx context.Context, id, status, processedPath, thumbPath string) error {
	const query = `
		UPDATE images
		SET status = $2, processed_path = $3, thumb_path = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, status, processedPath, thumbPath)
	return err
}

// DeleteImage удаляет запись об изображении.
func (r *PostgresRepository) DeleteImage(ctx context.Context, id string) error {
	const query = `DELETE FROM images WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrImageNotFound
	}

	return nil
}
