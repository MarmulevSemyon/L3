package service

import "errors"

// ErrForbidden возвращается, если у пользователя нет прав на действие.
var ErrForbidden = errors.New("forbidden")

// ErrInvalidInput возвращается, если входные данные некорректны.
var ErrInvalidInput = errors.New("invalid input")

// ErrNotFound возвращается, если запись не найдена.
var ErrNotFound = errors.New("not found")
