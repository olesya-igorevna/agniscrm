package handler

import (
	"agniscrm/internal/usecase"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct{ contacts *usecase.Contacts }

func New(store usecase.ContactRepository) *Handler {
	return &Handler{contacts: usecase.NewContacts(store)}
}

func (h *Handler) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("request panic: %v", recovered)
		writeError(c, 500, "internal_error", "внутренняя ошибка сервера")
	}))
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) { writeError(c, 404, "route_not_found", "путь не найден") })
	r.NoMethod(func(c *gin.Context) {
		writeError(c, 405, "method_not_allowed", "метод не поддерживается")
	})
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	h.RegisterContactRoutes(r.Group("/api/v1"))
	return r
}

// RegisterContactRoutes also works with the team's shared router.
func (h *Handler) RegisterContactRoutes(api *gin.RouterGroup) {
	c := api.Group("/contacts")
	c.GET("", h.ListContacts)
	c.GET("/:id", h.GetContact)
	c.POST("", h.CreateContact)
	c.POST("/:id/update", h.UpdateContact)
	c.POST("/:id/delete", h.DeleteContact)
}

func (h *Handler) Run(addr string) error {
	server := &http.Server{Addr: addr, Handler: h.Router(), ReadHeaderTimeout: 5 * time.Second}
	return server.ListenAndServe()
}
