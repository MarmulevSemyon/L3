package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"warehousecontrol/internal/domain"
	"warehousecontrol/internal/repository"
)

// ItemRepository описывает методы репозитория товаров.
type ItemRepository interface {
	Create(ctx context.Context, actor domain.Actor, input domain.CreateItemInput) (domain.Item, error)
	List(ctx context.Context) ([]domain.Item, error)
	GetByID(ctx context.Context, id int64) (domain.Item, error)
	Update(ctx context.Context, actor domain.Actor, id int64, input domain.UpdateItemInput) (domain.Item, error)
	Delete(ctx context.Context, actor domain.Actor, id int64) error
}

// ItemService содержит бизнес-логику работы с товарами.
type ItemService struct {
	repo ItemRepository
}

// NewItemService создаёт сервис товаров.
func NewItemService(repo ItemRepository) *ItemService {
	return &ItemService{
		repo: repo,
	}
}

// Create создаёт новый товар.
func (s *ItemService) Create(
	ctx context.Context,
	actor domain.Actor,
	input domain.CreateItemInput,
) (domain.Item, error) {
	if !canWriteItems(actor.Role) {
		return domain.Item{}, ErrForbidden
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if err := validateCreateItemInput(input); err != nil {
		return domain.Item{}, err
	}

	item, err := s.repo.Create(ctx, actor, input)
	if err != nil {
		return domain.Item{}, fmt.Errorf("create item: %w", err)
	}

	return item, nil
}

// List возвращает список товаров.
func (s *ItemService) List(ctx context.Context, actor domain.Actor) ([]domain.Item, error) {
	if !canReadItems(actor.Role) {
		return nil, ErrForbidden
	}

	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	return items, nil
}

// GetByID возвращает товар по ID.
func (s *ItemService) GetByID(ctx context.Context, actor domain.Actor, id int64) (domain.Item, error) {
	if !canReadItems(actor.Role) {
		return domain.Item{}, ErrForbidden
	}

	if id <= 0 {
		return domain.Item{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.Item{}, ErrNotFound
		}

		return domain.Item{}, fmt.Errorf("get item by id: %w", err)
	}

	return item, nil
}

// Update обновляет товар.
func (s *ItemService) Update(
	ctx context.Context,
	actor domain.Actor,
	id int64,
	input domain.UpdateItemInput,
) (domain.Item, error) {
	if !canWriteItems(actor.Role) {
		return domain.Item{}, ErrForbidden
	}

	if id <= 0 {
		return domain.Item{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if err := validateUpdateItemInput(input); err != nil {
		return domain.Item{}, err
	}

	item, err := s.repo.Update(ctx, actor, id, input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.Item{}, ErrNotFound
		}

		return domain.Item{}, fmt.Errorf("update item: %w", err)
	}

	return item, nil
}

// Delete удаляет товар.
func (s *ItemService) Delete(ctx context.Context, actor domain.Actor, id int64) error {
	if !canDeleteItems(actor.Role) {
		return ErrForbidden
	}

	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if err := s.repo.Delete(ctx, actor, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}

		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}

func validateCreateItemInput(input domain.CreateItemInput) error {
	if input.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	if input.Quantity < 0 {
		return fmt.Errorf("%w: quantity must be greater than or equal to zero", ErrInvalidInput)
	}

	if input.Price < 0 {
		return fmt.Errorf("%w: price must be greater than or equal to zero", ErrInvalidInput)
	}

	return nil
}

func validateUpdateItemInput(input domain.UpdateItemInput) error {
	if input.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	if input.Quantity < 0 {
		return fmt.Errorf("%w: quantity must be greater than or equal to zero", ErrInvalidInput)
	}

	if input.Price < 0 {
		return fmt.Errorf("%w: price must be greater than or equal to zero", ErrInvalidInput)
	}

	return nil
}

func canReadItems(role string) bool {
	switch role {
	case "admin", "manager", "viewer":
		return true
	default:
		return false
	}
}

func canWriteItems(role string) bool {
	switch role {
	case "admin", "manager":
		return true
	default:
		return false
	}
}

func canDeleteItems(role string) bool {
	return role == "admin"
}
