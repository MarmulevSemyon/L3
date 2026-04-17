package scheduler

import (
	"context"
	"time"

	"github.com/wb-go/wbf/logger"

	"eventBooker/internal/domain"
)

// BookingCanceller описывает сервис, умеющий отменять просроченные бронирования.
type BookingCanceller interface {
	CancelExpired(ctx context.Context) ([]*domain.Booking, error)
}

// Scheduler запускает фоновые задачи по расписанию.
type Scheduler struct {
	bookingService BookingCanceller
	interval       time.Duration
	logger         logger.Logger
}

// New создаёт новый экземпляр планировщика фоновых задач.
func New(bookingService BookingCanceller, interval time.Duration, logger logger.Logger) *Scheduler {
	return &Scheduler{
		bookingService: bookingService,
		interval:       interval,
		logger:         logger,
	}
}

// Start запускает цикл планировщика и периодически отменяет просроченные бронирования.
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.logger.Info("scheduler started", logger.Duration("interval", s.interval))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	cancelled, err := s.bookingService.CancelExpired(ctx)
	if err != nil {
		s.logger.Error("failed to cancel expired bookings", logger.String("error", err.Error()))
		return
	}

	for _, b := range cancelled {
		s.logger.Info(
			"booking expired",
			logger.String("booking_id", b.ID),
			logger.String("event_id", b.EventID),
			logger.String("user_id", b.UserID),
		)
	}
}
