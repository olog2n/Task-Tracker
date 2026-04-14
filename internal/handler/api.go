package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type ApiHandler struct{}

func (h *ApiHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("failed to encode response: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *ApiHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Issue Tracker API", "version": "0.0.1"}); err != nil {
		log.Printf("failed to encode response: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
