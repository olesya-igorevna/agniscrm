package handler

import (
        "encoding/json"
        "net/http"

        "agnis-crm/internal/usecase"
)

type Handler struct {
        uc *usecase.Usecase
}

func New(uc *usecase.Usecase) *Handler {
        return &Handler{uc: uc}
}

func (h *Handler) Register(mux *http.ServeMux) {
        mux.HandleFunc("GET /health", h.health)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
        if err := h.uc.Health(r.Context()); err != nil {
                writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "error"})
                return
        }
        writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.WriteHeader(status)
        _ = json.NewEncoder(w).Encode(v)
}