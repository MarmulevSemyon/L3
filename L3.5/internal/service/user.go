package service

import (
	"context"

	"eventBooker/internal/domain"
)

// UserRepo описывает методы репозитория пользователей, используемые сервисом.
type UserRepo interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
}

// UserService реализует бизнес-логику работы с пользователями.
type UserService struct {
	repo UserRepo
}

// NewUserService создаёт сервис для работы с пользователями.
func NewUserService(repo UserRepo) *UserService {
	return &UserService{repo: repo}
}

// Create создаёт нового пользователя.
func (s *UserService) Create(ctx context.Context, u *domain.User) error {
	return s.repo.Create(ctx, u)
}

// List возвращает список всех пользователей.
func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
	return s.repo.List(ctx)
}

// GetByID возвращает пользователя по его идентификатору.
func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}
