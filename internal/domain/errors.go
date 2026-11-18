package domain

import "errors"

var (
	ErrNotFound    = errors.New("entity not found")
	ErrConflict    = errors.New("entity is in use")
	ErrTimeout     = errors.New("operation timed out")
	ErrUnavailable = errors.New("storage unavailable")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Details []FieldError
}

func (e *ValidationError) Error() string { return "некорректные данные" }
