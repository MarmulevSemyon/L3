package service

import (
	"context"

	"eventBooker/internal/domain"
)

// EventRepo описывает методы репозитория мероприятий, используемые сервисом.
type EventRepo interface {
	Create(ctx context.Context, e *domain.Event) error
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	List(ctx context.Context) ([]*domain.Event, error)
	GetDetails(ctx context.Context, eventID string) (*domain.EventDetails, error)
}

// BookingEventRepo описывает методы репозитория бронирований, необходимые сервису мероприятий.
type BookingEventRepo interface {
	ListByEvent(ctx context.Context, eventID string) ([]*domain.Booking, error)
}

// EventService реализует бизнес-логику работы с мероприятиями.
type EventService struct {
	eventRepo   EventRepo
	bookingRepo BookingEventRepo
}

// NewEventService создаёт сервис для работы с мероприятиями.
func NewEventService(eventRepo EventRepo, bookingRepo BookingEventRepo) *EventService {
	return &EventService{
		eventRepo:   eventRepo,
		bookingRepo: bookingRepo,
	}
}

// Create создаёт новое мероприятие.
func (s *EventService) Create(ctx context.Context, e *domain.Event) error {
	return s.eventRepo.Create(ctx, e)
}

// GetDetails возвращает подробную информацию о мероприятии и его бронированиях.
func (s *EventService) GetDetails(ctx context.Context, id string) (*domain.EventDetails, error) {
	details, err := s.eventRepo.GetDetails(ctx, id)
	if err != nil {
		return nil, err
	}

	bookings, err := s.bookingRepo.ListByEvent(ctx, id)
	if err != nil {
		return nil, err
	}

	details.Bookings = bookings
	return details, nil
}

// List возвращает список всех мероприятий.
func (s *EventService) List(ctx context.Context) ([]*domain.Event, error) {
	return s.eventRepo.List(ctx)
}
