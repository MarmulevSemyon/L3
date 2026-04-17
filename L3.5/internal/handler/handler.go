package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"

	"eventBooker/internal/domain"
)

// EventService описывает методы сервиса мероприятий, используемые обработчиками HTTP-запросов.
type EventService interface {
	Create(ctx context.Context, e *domain.Event) error
	GetDetails(ctx context.Context, id string) (*domain.EventDetails, error)
	List(ctx context.Context) ([]*domain.Event, error)
}

// BookingService описывает методы сервиса бронирований, используемые обработчиками HTTP-запросов.
type BookingService interface {
	Book(ctx context.Context, eventID, userID string) error
	Confirm(ctx context.Context, eventID, userID string) error
	ListByUser(ctx context.Context, userID string) ([]*domain.Booking, error)
}

// UserService описывает методы сервиса пользователей, используемые обработчиками HTTP-запросов.
type UserService interface {
	Create(ctx context.Context, u *domain.User) error
	List(ctx context.Context) ([]*domain.User, error)
}

// Handler объединяет HTTP-обработчики приложения.
type Handler struct {
	eventService   EventService
	bookingService BookingService
	userService    UserService
}

// NewHandler создаёт новый набор HTTP-обработчиков приложения.
func NewHandler(es EventService, bs BookingService, us UserService) *Handler {
	return &Handler{
		eventService:   es,
		bookingService: bs,
		userService:    us,
	}
}

type createEventRequest struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	EventDate       time.Time `json:"event_date"`
	TotalSpots      int       `json:"total_spots"`
	RequiresPayment bool      `json:"requires_payment"`
	BookingTTLMin   int       `json:"booking_ttl_min"`
}

// CreateEvent обрабатывает HTTP-запрос на создание мероприятия.
func (h *Handler) CreateEvent(c *ginext.Context) {
	var req createEventRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	e := &domain.Event{
		ID:              uuid.NewString(),
		Title:           req.Title,
		Description:     req.Description,
		EventDate:       req.EventDate,
		TotalSpots:      req.TotalSpots,
		RequiresPayment: req.RequiresPayment,
		BookingTTL:      time.Duration(req.BookingTTLMin) * time.Minute,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.eventService.Create(c.Request.Context(), e); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{"event": e})
}

// ListEvents обрабатывает HTTP-запрос на получение списка мероприятий.
func (h *Handler) ListEvents(c *ginext.Context) {
	events, err := h.eventService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"events": events})
}

// GetEvent обрабатывает HTTP-запрос на получение подробной информации о мероприятии.
func (h *Handler) GetEvent(c *ginext.Context) {
	id := c.Param("id")

	event, err := h.eventService.GetDetails(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrEventNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"event": event})
}

type bookingRequest struct {
	UserID string `json:"user_id"`
}

// BookEvent обрабатывает HTTP-запрос на бронирование места на мероприятии.
func (h *Handler) BookEvent(c *ginext.Context) {
	eventID := c.Param("id")

	var req bookingRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	if err := h.bookingService.Book(c.Request.Context(), eventID, req.UserID); err != nil {
		status := http.StatusInternalServerError
		switch err {
		case domain.ErrNoAvailableSpots, domain.ErrAlreadyBooked:
			status = http.StatusConflict
		case domain.ErrEventNotFound, domain.ErrUserNotFound:
			status = http.StatusNotFound
		}
		c.JSON(status, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{"message": "booking created"})
}

// ConfirmBooking обрабатывает HTTP-запрос на подтверждение бронирования.
func (h *Handler) ConfirmBooking(c *ginext.Context) {
	eventID := c.Param("id")

	var req bookingRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	if err := h.bookingService.Confirm(c.Request.Context(), eventID, req.UserID); err != nil {
		status := http.StatusInternalServerError
		switch err {
		case domain.ErrBookingExpired, domain.ErrBookingNotPending:
			status = http.StatusConflict
		case domain.ErrBookingNotFound, domain.ErrEventNotFound:
			status = http.StatusNotFound
		}
		c.JSON(status, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "booking confirmed"})
}

type createUserRequest struct {
	Username string `json:"username"`
}

// CreateUser обрабатывает HTTP-запрос на создание пользователя.
func (h *Handler) CreateUser(c *ginext.Context) {
	var req createUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	u := &domain.User{
		ID:        uuid.NewString(),
		Username:  req.Username,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.userService.Create(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{"user": u})
}

// ListUsers обрабатывает HTTP-запрос на получение списка пользователей.
func (h *Handler) ListUsers(c *ginext.Context) {
	users, err := h.userService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"users": users})
}

// GetUserBookings обрабатывает HTTP-запрос на получение списка бронирований пользователя.
func (h *Handler) GetUserBookings(c *ginext.Context) {
	userID := c.Param("id")

	bookings, err := h.bookingService.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"bookings": bookings})
}
