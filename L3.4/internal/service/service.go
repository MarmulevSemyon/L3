package service

import (
	"context"
	"fmt"
	"imageProcessor/internal/broker"
	"imageProcessor/internal/dto"
	"imageProcessor/internal/model"
	"imageProcessor/internal/repository"
	"imageProcessor/internal/storage"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Service описывает бизнес-логику сервиса.
type Service struct {
	repo     repository.Repository
	storage  *storage.FileStore
	producer broker.Producer
	baseURL  string
}

// Processor описывает обработчик Kafka-сообщений.
type Processor interface {
	ProcessMessage(ctx context.Context, msg dto.ProcessImageMessage) error
}

// New создает сервис.
func New(
	repo repository.Repository,
	storage *storage.FileStore,
	producer broker.Producer,
) *Service {
	return &Service{
		repo:     repo,
		storage:  storage,
		producer: producer,
	}
}

// UploadImage сохраняет файл и ставит задачу в очередь.
func (s *Service) UploadImage(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
	resize, thumb, watermark bool,
) (*dto.UploadResponse, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		return nil, fmt.Errorf("empty file extension")
	}

	id := uuid.New().String()
	filename := id + ext

	originalPath, err := s.storage.SaveOriginal(filename, file)
	if err != nil {
		return nil, err
	}

	image := &model.Image{
		ID:                 id,
		OriginalName:       header.Filename,
		ContentType:        header.Header.Get("Content-Type"),
		Status:             "pending",
		OperationResize:    resize,
		OperationThumb:     thumb,
		OperationWatermark: watermark,
		OriginalPath:       originalPath,
	}

	if err := s.repo.CreateImage(ctx, image); err != nil {
		_ = s.storage.DeleteFile(originalPath)
		return nil, err
	}

	msg := dto.ProcessImageMessage{
		ID:           id,
		OriginalPath: originalPath,
		Resize:       resize,
		Thumb:        thumb,
		Watermark:    watermark,
	}

	if err := s.producer.Send(ctx, msg); err != nil {
		_ = s.storage.DeleteFile(originalPath)
		_ = s.repo.DeleteImage(ctx, id)
		return nil, err
	}

	return &dto.UploadResponse{
		ID:     id,
		Status: "pending",
	}, nil
}

// GetImageStatus возвращает состояние изображения.
func (s *Service) GetImageStatus(ctx context.Context, id string) (*dto.ImageStatusResponse, error) {
	image, err := s.repo.GetImageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &dto.ImageStatusResponse{
		ID:        image.ID,
		Status:    image.Status,
		ErrorText: image.ErrorText,
	}

	if image.ProcessedPath != "" {
		resp.ImageURL = "/files/processed/" + filepath.Base(image.ProcessedPath)
	}
	if image.ThumbPath != "" {
		resp.ThumbURL = "/files/thumbs/" + filepath.Base(image.ThumbPath)
	}

	return resp, nil
}

// DeleteImage удаляет запись и файлы.
func (s *Service) DeleteImage(ctx context.Context, id string) error {
	image, err := s.repo.GetImageByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.storage.DeleteFile(image.OriginalPath); err != nil {
		return err
	}
	if err := s.storage.DeleteFile(image.ProcessedPath); err != nil {
		return err
	}
	if err := s.storage.DeleteFile(image.ThumbPath); err != nil {
		return err
	}

	return s.repo.DeleteImage(ctx, id)
}
