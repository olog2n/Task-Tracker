package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// respondJSON — удобный хелпер для JSON-ответов
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("failed to encode response: %v", err)
		}
	}
}

// respondError — хелпер для ошибок
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{"error": message})
}

func ParseUUIDParam(r *http.Request, paramName string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("%s is required", paramName)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s: %w", paramName, err)
	}

	return id, nil
}

// parseUUIDQuery — парсит UUID из query param
func ParseUUIDQuery(r *http.Request, paramName string) (*uuid.UUID, error) {
	val := r.URL.Query().Get(paramName)
	if val == "" {
		return nil, nil // Не ошибка, просто не указан
	}

	id, err := uuid.Parse(val)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", paramName, err)
	}

	return &id, nil
}
