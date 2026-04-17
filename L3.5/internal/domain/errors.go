package domain

import "errors"

var (
	// ErrEventNotFound означает, что мероприятие не найдено.
	ErrEventNotFound = errors.New("event not found")

	// ErrUserNotFound означает, что пользователь не найден.
	ErrUserNotFound = errors.New("user not found")

	// ErrBookingNotFound означает, что бронирование не найдено.
	ErrBookingNotFound = errors.New("booking not found")

	// ErrNoAvailableSpots означает, что свободных мест на мероприятии больше нет.
	ErrNoAvailableSpots = errors.New("no available spots")

	// ErrAlreadyBooked означает, что пользователь уже имеет активное бронирование на это мероприятие.
	ErrAlreadyBooked = errors.New("user already has active booking for this event")

	// ErrBookingExpired означает, что время жизни бронирования истекло.
	ErrBookingExpired = errors.New("booking expired")

	// ErrBookingNotPending означает, что бронирование находится не в статусе ожидания подтверждения.
	ErrBookingNotPending = errors.New("booking is not pending")

	// ErrInvalidInput означает, что входные данные некорректны.
	ErrInvalidInput = errors.New("invalid input")
)
