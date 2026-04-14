package repository

import (
	"context"
	"imageProcessor/internal/model"
)

// Repository описывает методы работы с хранилищем метаданных изображений.
type Repository interface {
	CreateImage(ctx context.Context, image *model.Image) error
	GetImageByID(ctx context.Context, id string) (*model.Image, error)
	UpdateImageStatus(ctx context.Context, id string, status string, errText string) error
	UpdateImageResult(ctx context.Context, id, status, processedPath, thumbPath string) error
	DeleteImage(ctx context.Context, id string) error
}
