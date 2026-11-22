package handler

import (
	"agnis-crm/internal/domain"
	"agnis-crm/internal/httpx"
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

var writeError = httpx.WriteError

func domainError(c *gin.Context, err error) {
	var validation *domain.ValidationError
	switch {
	case errors.As(err, &validation):
		writeError(c, 400, "validation_error", "некорректные данные", validation.Details...)
	case errors.Is(err, domain.ErrNotFound):
		writeError(c, 404, "not_found", "запись не найдена")
	case errors.Is(err, domain.ErrConflict):
		writeError(c, 409, "conflict", "запись используется в других данных")
	default:
		log.Printf("operation failed: %v", err)
		writeError(c, 500, "internal_error", "внутренняя ошибка сервера")
	}
}
