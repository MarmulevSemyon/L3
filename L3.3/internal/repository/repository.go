package repository

import (
	"context"

	"commentTree/internal/model"
)

// CommentRepository описывает работу с хранилищем комментариев.
type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByID(ctx context.Context, id int64) (*model.Comment, error)
	GetChildren(ctx context.Context, parentID *int64, limit, offset int, sortBy, order string) ([]model.Comment, int, error)
	GetAllRoots(ctx context.Context, limit, offset int, sortBy, order string) ([]model.Comment, int, error)
	DeleteSubtree(ctx context.Context, id int64) error
	Search(ctx context.Context, query string, limit, offset int, sortBy, order string) ([]model.Comment, int, error)
}
