package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"salestracker/internal/domain"
	"salestracker/internal/dto"
	"salestracker/internal/repository/postgres"
)

var (
	// ErrInvalidType означает, что передан неверный тип записи.
	ErrInvalidType = errors.New("invalid type")
	// ErrInvalidAmount означает, что сумма отрицательная.
	ErrInvalidAmount = errors.New("amount must be non-negative")
	// ErrInvalidCategory означает, что категория пустая.
	ErrInvalidCategory = errors.New("category is required")
	// ErrInvalidOccurredAt означает, что дата отсутствует.
	ErrInvalidOccurredAt = errors.New("occurred_at is required")
)

// ItemService содержит бизнес-логику работы с записями.
type ItemService struct {
	repo *postgres.ItemRepository
}

// NewItemService создаёт сервис записей.
func NewItemService(repo *postgres.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

// Create создаёт запись.
func (s *ItemService) Create(ctx context.Context, req dto.CreateItemRequest) (*domain.Item, error) {
	if err := validate(req.Type, req.Amount, req.Category, req.OccurredAt); err != nil {
		return nil, err
	}

	item := &domain.Item{
		Type:        req.Type,
		Amount:      req.Amount,
		Category:    strings.TrimSpace(req.Category),
		Description: strings.TrimSpace(req.Description),
		OccurredAt:  req.OccurredAt,
	}

	return s.repo.Create(ctx, item)
}

// GetByID возвращает запись по ID.
func (s *ItemService) GetByID(ctx context.Context, id int64) (*domain.Item, error) {
	return s.repo.GetByID(ctx, id)
}

// List возвращает список записей.
func (s *ItemService) List(ctx context.Context, filter domain.ListFilter) ([]domain.Item, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if filter.SortBy == "" {
		filter.SortBy = "occurred_at"
	}
	if filter.Order == "" {
		filter.Order = "desc"
	}

	return s.repo.List(ctx, filter)
}

// Update обновляет запись.
func (s *ItemService) Update(ctx context.Context, id int64, req dto.UpdateItemRequest) (*domain.Item, error) {
	if err := validate(req.Type, req.Amount, req.Category, req.OccurredAt); err != nil {
		return nil, err
	}

	item := &domain.Item{
		Type:        req.Type,
		Amount:      req.Amount,
		Category:    strings.TrimSpace(req.Category),
		Description: strings.TrimSpace(req.Description),
		OccurredAt:  req.OccurredAt,
	}

	return s.repo.Update(ctx, id, item)
}

// Delete удаляет запись.
func (s *ItemService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// Analytics возвращает агрегированную аналитику.
func (s *ItemService) Analytics(ctx context.Context, filter domain.AnalyticsFilter) (*domain.Analytics, error) {
	return s.repo.Analytics(ctx, filter)
}

func validate(itemType string, amount float64, category string, occurredAt time.Time) error {
	if itemType != "income" && itemType != "expense" {
		return ErrInvalidType
	}
	if amount < 0 {
		return ErrInvalidAmount
	}
	if strings.TrimSpace(category) == "" {
		return ErrInvalidCategory
	}
	if occurredAt.IsZero() {
		return ErrInvalidOccurredAt
	}

	return nil
}
