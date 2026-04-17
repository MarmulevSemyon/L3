package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"

	"eventBooker/internal/domain"
)

// BookingRepository предоставляет методы работы с бронированиями в базе данных.
type BookingRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

// NewBookingRepo создаёт репозиторий для работы с бронированиями.
func NewBookingRepo(db *dbpg.DB) *BookingRepository {
	return &BookingRepository{
		db: db,
		strategy: retry.Strategy{
			Attempts: 3,
			Delay:    500 * time.Millisecond,
			Backoff:  2,
		},
	}
}

// Create создаёт новое бронирование в базе данных с учётом ограничения по количеству мест.
func (r *BookingRepository) Create(ctx context.Context, b *domain.Booking) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var totalSpots int
	err = tx.QueryRowContext(ctx,
		`SELECT total_spots FROM events WHERE id = $1 FOR UPDATE`,
		b.EventID,
	).Scan(&totalSpots)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrEventNotFound
		}
		return fmt.Errorf("get total spots: %w", err)
	}

	var activeBookings int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM bookings WHERE event_id = $1 AND status = ANY($2)`,
		b.EventID,
		pq.Array(domain.ActiveStatuses),
	).Scan(&activeBookings)
	if err != nil {
		return fmt.Errorf("count active bookings: %w", err)
	}

	if activeBookings >= totalSpots {
		return domain.ErrNoAvailableSpots
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO bookings (id, event_id, user_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		b.ID, b.EventID, b.UserID, b.Status, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyBooked
		}
		return fmt.Errorf("insert booking: %w", err)
	}

	return tx.Commit()
}

// GetByEventAndUser возвращает активное бронирование пользователя для указанного мероприятия.
func (r *BookingRepository) GetByEventAndUser(ctx context.Context, eventID, userID string) (*domain.Booking, error) {
	query := `
		SELECT id, event_id, user_id, status, created_at, updated_at
		FROM bookings
		WHERE event_id = $1
		  AND user_id = $2
		  AND status = ANY($3)
		ORDER BY created_at DESC
		LIMIT 1
	`

	row, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, eventID, userID, pq.Array(domain.ActiveStatuses))
	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}

	var b domain.Booking
	if err = row.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, fmt.Errorf("scan booking: %w", err)
	}

	return &b, nil
}

// ListByUser возвращает список бронирований указанного пользователя.
func (r *BookingRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Booking, error) {
	query := `
		SELECT id, event_id, user_id, status, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list bookings by user: %w", err)
	}
	defer rows.Close()

	var res []*domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err = rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		res = append(res, &b)
	}

	return res, rows.Err()
}

// ListByEvent возвращает список активных бронирований для указанного мероприятия.
func (r *BookingRepository) ListByEvent(ctx context.Context, eventID string) ([]*domain.Booking, error) {
	query := `
		SELECT id, event_id, user_id, status, created_at, updated_at
		FROM bookings
		WHERE event_id = $1
		  AND status = ANY($2)
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query, eventID, pq.Array(domain.ActiveStatuses))
	if err != nil {
		return nil, fmt.Errorf("list bookings by event: %w", err)
	}
	defer rows.Close()

	var res []*domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err = rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		res = append(res, &b)
	}

	return res, rows.Err()
}

// Confirm подтверждает бронирование пользователя, если срок его действия ещё не истёк.
func (r *BookingRepository) Confirm(ctx context.Context, eventID, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var ttlSeconds int64
	err = tx.QueryRowContext(ctx,
		`SELECT EXTRACT(EPOCH FROM booking_ttl)::bigint FROM events WHERE id = $1`,
		eventID,
	).Scan(&ttlSeconds)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrEventNotFound
		}
		return fmt.Errorf("get event ttl: %w", err)
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE bookings
		 SET status = $4, updated_at = NOW()
		 WHERE event_id = $1
		   AND user_id = $2
		   AND status = $3
		   AND created_at + make_interval(secs => $5) >= NOW()`,
		eventID,
		userID,
		domain.BookingStatusPending,
		domain.BookingStatusConfirmed,
		ttlSeconds,
	)
	if err != nil {
		return fmt.Errorf("confirm booking: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		var status string
		var createdAt time.Time

		scanErr := tx.QueryRowContext(ctx,
			`SELECT status, created_at
			 FROM bookings
			 WHERE event_id = $1
			   AND user_id = $2
			   AND status = ANY($3)
			 ORDER BY created_at DESC
			 LIMIT 1`,
			eventID,
			userID,
			pq.Array(domain.ActiveStatuses),
		).Scan(&status, &createdAt)

		if scanErr != nil {
			return domain.ErrBookingNotFound
		}
		if status != string(domain.BookingStatusPending) {
			return domain.ErrBookingNotPending
		}
		if time.Since(createdAt) > time.Duration(ttlSeconds)*time.Second {
			return domain.ErrBookingExpired
		}
		return domain.ErrBookingNotFound
	}

	return tx.Commit()
}

// CancelExpired отменяет все просроченные неподтверждённые бронирования.
func (r *BookingRepository) CancelExpired(ctx context.Context) ([]*domain.Booking, error) {
	query := `
		UPDATE bookings b
		SET status = $2, updated_at = NOW()
		FROM events e
		WHERE b.event_id = e.id
		  AND b.status = $1
		  AND b.created_at + e.booking_ttl < NOW()
		RETURNING b.id, b.event_id, b.user_id, b.status, b.created_at, b.updated_at
	`

	rows, err := r.db.QueryWithRetry(
		ctx, r.strategy, query,
		domain.BookingStatusPending,
		domain.BookingStatusCancelled,
	)
	if err != nil {
		return nil, fmt.Errorf("cancel expired: %w", err)
	}
	defer rows.Close()

	var res []*domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err = rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan cancelled booking: %w", err)
		}
		res = append(res, &b)
	}

	return res, rows.Err()
}
