package handler

import (
	"agniscrm/internal/domain"
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, status int, code, message string, details ...domain.FieldError) {
	body := gin.H{"code": code, "message": message}
	if len(details) > 0 {
		body["details"] = details
	}
	c.AbortWithStatusJSON(status, gin.H{"error": body})
}

func domainError(c *gin.Context, err error) {
	var validation *domain.ValidationError
	switch {
	case errors.As(err, &validation):
		writeError(c, 400, "validation_error", "некорректные данные", validation.Details...)
	case errors.Is(err, domain.ErrNotFound):
		writeError(c, 404, "not_found", "контакт не найден")
	case errors.Is(err, domain.ErrConflict):
		writeError(c, 409, "conflict", "контакт используется в сделках или задачах")
	default:
		log.Printf("contact operation failed: %v", err)
		writeError(c, 500, "internal_error", "внутренняя ошибка сервера")
	}
}
