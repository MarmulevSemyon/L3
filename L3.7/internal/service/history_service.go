package service

import (
	"context"
	"fmt"

	"warehousecontrol/internal/domain"
)

// HistoryRepository описывает методы репозитория истории.
type HistoryRepository interface {
	ListByItemID(ctx context.Context, itemID int64) ([]domain.ItemHistory, error)
}

// HistoryService содержит бизнес-логику просмотра истории.
type HistoryService struct {
	repo HistoryRepository
}

// NewHistoryService создаёт сервис истории.
func NewHistoryService(repo HistoryRepository) *HistoryService {
	return &HistoryService{
		repo: repo,
	}
}

// ListByItemID возвращает историю изменений товара.
func (s *HistoryService) ListByItemID(
	ctx context.Context,
	actor domain.Actor,
	itemID int64,
) ([]domain.ItemHistory, error) {
	if !canReadItems(actor.Role) {
		return nil, ErrForbidden
	}

	if itemID <= 0 {
		return nil, fmt.Errorf("%w: item id must be positive", ErrInvalidInput)
	}

	history, err := s.repo.ListByItemID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("list item history: %w", err)
	}

	return history, nil
}
