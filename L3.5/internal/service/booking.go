package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/logger"

	"eventBooker/internal/domain"
)

// BookingRepo описывает методы репозитория бронирований, используемые сервисом.
type BookingRepo interface {
	Create(ctx context.Context, b *domain.Booking) error
	Confirm(ctx context.Context, eventID, userID string) error
	CancelExpired(ctx context.Context) ([]*domain.Booking, error)
	ListByUser(ctx context.Context, userID string) ([]*domain.Booking, error)
}

// EventBookingRepo описывает методы репозитория мероприятий, необходимые сервису бронирований.
type EventBookingRepo interface {
	GetByID(ctx context.Context, id string) (*domain.Event, error)
}

// UserBookingRepo описывает методы репозитория пользователей, необходимые сервису бронирований.
type UserBookingRepo interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

// BookingService реализует бизнес-логику бронирования мест на мероприятия.
type BookingService struct {
	bookingRepo BookingRepo
	eventRepo   EventBookingRepo
	userRepo    UserBookingRepo
	log         logger.Logger
}

// NewBookingService создаёт сервис для работы с бронированиями.
func NewBookingService(
	bookingRepo BookingRepo,
	eventRepo EventBookingRepo,
	userRepo UserBookingRepo,
	log logger.Logger,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		eventRepo:   eventRepo,
		userRepo:    userRepo,
		log:         log,
	}
}

// Book создаёт новое бронирование пользователя на мероприятие.
func (s *BookingService) Book(ctx context.Context, eventID, userID string) error {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	if _, err = s.userRepo.GetByID(ctx, userID); err != nil {
		return err
	}

	status := domain.BookingStatusPending
	if !event.RequiresPayment {
		status = domain.BookingStatusConfirmed
	}

	now := time.Now().UTC()
	booking := &domain.Booking{
		ID:        uuid.NewString(),
		EventID:   eventID,
		UserID:    userID,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.bookingRepo.Create(ctx, booking)
}

// Confirm подтверждает бронирование пользователя на мероприятие.
func (s *BookingService) Confirm(ctx context.Context, eventID, userID string) error {
	return s.bookingRepo.Confirm(ctx, eventID, userID)
}

// CancelExpired отменяет все просроченные бронирования.
func (s *BookingService) CancelExpired(ctx context.Context) ([]*domain.Booking, error) {
	return s.bookingRepo.CancelExpired(ctx)
}

// ListByUser возвращает список бронирований указанного пользователя.
func (s *BookingService) ListByUser(ctx context.Context, userID string) ([]*domain.Booking, error) {
	return s.bookingRepo.ListByUser(ctx, userID)
}
