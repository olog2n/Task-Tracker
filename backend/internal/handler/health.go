package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct {
	metadata *MetadataService
	db       interface{} //TODO add repository_health
}

func NewHealthHandler(metadata *MetadataService, db interface{}) *HealthHandler {
	return &HealthHandler{
		metadata: metadata,
		db:       db,
	}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    h.metadata.Uptime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// if err := h.db.Ping(); err != nil {
	//     w.WriteHeader(http.StatusServiceUnavailable)
	//     return
	// }

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    h.metadata.Uptime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
