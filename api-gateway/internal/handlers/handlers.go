package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"bike-rental/api-gateway/internal/client"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	rentClient client.RentClient
	statsClient client.StatsClient
}

func NewHandlers(rentClient client.RentClient, statsClient client.StatsClient) *Handlers {
	return &Handlers{
		rentClient:  rentClient,
		statsClient: statsClient,
	}
}

// @Summary Start a bike rent
// @Description Start renting a bike
// @Tags rent
// @Accept json
// @Produce json
// @Param request body StartRentRequest true "Start rent request"
// @Success 200 {object} RentResponse
// @Router /api/v1/rent/start [post]
func (h *Handlers) StartRent(w http.ResponseWriter, r *http.Request) {
	var req StartRentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log request
	log.Printf("API Gateway: Received StartRent request: user_id=%s, bike_id=%s", req.UserID, req.BikeID)

	response, err := h.rentClient.StartRent(r.Context(), req.UserID, req.BikeID)
	if err != nil {
		log.Printf("API Gateway: Error calling rent service: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("API Gateway: StartRent successful: rent_id=%s", response.RentID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary End a bike rent
// @Description End an active bike rent
// @Tags rent
// @Accept json
// @Produce json
// @Param request body EndRentRequest true "End rent request"
// @Success 200 {object} RentResponse
// @Router /api/v1/rent/end [post]
func (h *Handlers) EndRent(w http.ResponseWriter, r *http.Request) {
	var req EndRentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.rentClient.EndRent(r.Context(), req.RentID, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get available bikes
// @Description Get list of available bikes
// @Tags bikes
// @Produce json
// @Param location query string false "Filter by location"
// @Success 200 {object} BikesListResponse
// @Router /api/v1/bikes/available [get]
func (h *Handlers) GetAvailableBikes(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")

	bikes, err := h.rentClient.GetAvailableBikes(r.Context(), location)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bikes)
}

// @Summary Get daily statistics
// @Description Get statistics for a specific date
// @Tags stats
// @Produce json
// @Param date path string true "Date in YYYY-MM-DD format"
// @Success 200 {object} DailyStatsResponse
// @Router /api/v1/stats/daily/{date} [get]
func (h *Handlers) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	date := chi.URLParam(r, "date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	stats, err := h.statsClient.GetDailyStats(r.Context(), date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// @Summary Get active rents count
// @Description Get current number of active rents
// @Tags stats
// @Produce json
// @Success 200 {object} ActiveRentsResponse
// @Router /api/v1/stats/active [get]
func (h *Handlers) GetActiveRents(w http.ResponseWriter, r *http.Request) {
	count, err := h.statsClient.GetActiveRents(r.Context())
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

// @Summary Health check
// @Description Health check endpoint
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handlers) RegisterRoutes(r *chi.Mux) {
	r.Post("/api/v1/rent/start", h.StartRent)
	r.Post("/api/v1/rent/end", h.EndRent)
	r.Get("/api/v1/bikes/available", h.GetAvailableBikes)
	r.Get("/api/v1/stats/daily/{date}", h.GetDailyStats)
	r.Get("/api/v1/stats/active", h.GetActiveRents)
	r.Get("/health", h.Health)
}

// Request/Response types
type StartRentRequest struct {
	UserID string `json:"user_id"`
	BikeID string `json:"bike_id"`
}

type EndRentRequest struct {
	RentID string `json:"rent_id"`
	UserID string `json:"user_id"`
}

type RentResponse struct {
	RentID    string `json:"rent_id"`
	UserID    string `json:"user_id"`
	BikeID    string `json:"bike_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

type BikesListResponse struct {
	Bikes []Bike `json:"bikes"`
}

type Bike struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Location string `json:"location"`
}

type DailyStatsResponse struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ActiveRentsResponse struct {
	ActiveRents int64 `json:"active_rents"`
}

