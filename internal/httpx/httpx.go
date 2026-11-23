package httpx

import (
	"github.com/gin-gonic/gin"

	"agnis-crm/internal/domain"
)

// WriteError отправляет ошибку в едином формате ТЗ и прерывает обработку запроса.
func WriteError(c *gin.Context, status int, code, message string, details ...domain.FieldError) {
	body := gin.H{"code": code, "message": message}
	if len(details) > 0 {
		body["details"] = details
	}
	c.AbortWithStatusJSON(status, gin.H{"error": body})
}
