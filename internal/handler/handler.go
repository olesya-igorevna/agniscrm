package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"agnis-crm/internal/usecase"
)

type Handler struct{ uc *usecase.Usecase }

func New(uc *usecase.Usecase) *Handler { return &Handler{uc: uc} }

// Logging — middleware логирования: метод, путь, код ответа, длительность (НФТ-08).
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}

func (h *Handler) Register(r *gin.Engine) {
	r.Use(Logging())
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("request panic: %v", recovered)
		writeError(c, 500, "internal_error", "внутренняя ошибка сервера")
	}))
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) { writeError(c, 404, "route_not_found", "путь не найден") })
	r.NoMethod(func(c *gin.Context) {
		writeError(c, 405, "method_not_allowed", "метод не поддерживается")
	})
	r.GET("/health", h.health)

	contacts := r.Group("/api/v1/contacts")
	contacts.GET("", h.ListContacts)
	contacts.GET("/:id", h.GetContact)
	contacts.POST("", h.CreateContact)
	contacts.POST("/:id/update", h.UpdateContact)
	contacts.POST("/:id/delete", h.DeleteContact)
}

func (h *Handler) health(c *gin.Context) {
	if err := h.uc.Health(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
