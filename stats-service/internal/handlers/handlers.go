package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Domenick1991/students/stats-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	service service.Service
}

func NewHandlers(svc service.Service) *Handlers {
	return &Handlers{service: svc}
}

func (h *Handlers) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	count, err := h.service.GetDailyStats(r.Context(), date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"date": date,
		"count": count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) GetActiveRents(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.GetActiveRents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"active_rents": count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) RefreshStats(w http.ResponseWriter, r *http.Request) {
	// This endpoint could trigger a recalculation of stats
	// For now, just return success
	response := map[string]string{
		"status": "ok",
		"message": "Stats refresh initiated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) RegisterRoutes(r *chi.Mux) {
	r.Get("/internal/stats/daily", h.GetDailyStats)
	r.Get("/internal/stats/active", h.GetActiveRents)
	r.Post("/admin/refresh-stats", h.RefreshStats)
}

