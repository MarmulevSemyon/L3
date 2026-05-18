package service

import (
	"context"
	"fmt"
	"strings"

	"warehousecontrol/internal/auth"
	"warehousecontrol/internal/domain"
)

// UserRepository описывает методы репозитория пользователей.
type UserRepository interface {
	GetActorByUsername(ctx context.Context, username string) (domain.Actor, error)
}

// AuthService содержит логику авторизации.
type AuthService struct {
	userRepo  UserRepository
	jwtSecret string
}

// NewAuthService создаёт сервис авторизации.
func NewAuthService(userRepo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// LoginRequest описывает данные для входа.
type LoginRequest struct {
	Username string `json:"username"`
}

// LoginResponse описывает ответ после входа.
type LoginResponse struct {
	Token string       `json:"token"`
	User  domain.Actor `json:"user"`
}

// Login проверяет пользователя и выдаёт JWT.
func (s *AuthService) Login(ctx context.Context, input LoginRequest) (LoginResponse, error) {
	username := strings.TrimSpace(input.Username)
	if username == "" {
		return LoginResponse{}, fmt.Errorf("%w: username is required", ErrInvalidInput)
	}

	actor, err := s.userRepo.GetActorByUsername(ctx, username)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("get user: %w", err)
	}

	token, err := auth.GenerateToken(actor, s.jwtSecret)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("generate token: %w", err)
	}

	return LoginResponse{
		Token: token,
		User:  actor,
	}, nil
}
