package domain

import "errors"

var (
	ErrNotFound = errors.New("entity not found")
	ErrConflict = errors.New("entity is in use")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Details []FieldError
}

func (e *ValidationError) Error() string { return "некорректные данные" }
