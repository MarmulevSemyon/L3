package domain

import "time"

// BookingStatus описывает текущее состояние бронирования.
type BookingStatus string

const (
	// BookingStatusPending означает, что бронирование создано, но ещё не подтверждено.
	BookingStatusPending BookingStatus = "pending"

	// BookingStatusConfirmed означает, что бронирование подтверждено.
	BookingStatusConfirmed BookingStatus = "confirmed"

	// BookingStatusCancelled означает, что бронирование отменено.
	BookingStatusCancelled BookingStatus = "cancelled"
)

// ActiveStatuses содержит статусы бронирований, которые считаются активными.
var ActiveStatuses = []string{
	string(BookingStatusPending),
	string(BookingStatusConfirmed),
}

// Booking описывает бронирование места пользователем на мероприятие.
type Booking struct {
	ID        string        `json:"id"`
	EventID   string        `json:"event_id"`
	UserID    string        `json:"user_id"`
	Status    BookingStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
