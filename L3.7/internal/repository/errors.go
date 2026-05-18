package repository

import "errors"

// ErrNotFound возвращается, если запись не найдена.
var ErrNotFound = errors.New("not found")
