package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"tracker/internal/auth"
	"tracker/internal/model"
	"tracker/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo repository.UserRepository
	jwt      *auth.JWTService
}

func NewAuthHandler(userRepo repository.UserRepository, jwt *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		jwt:      jwt,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input model.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if len(input.Password) < 6 {
		http.Error(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("hash error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := &model.User{
		Email:        input.Email,
		PasswordHash: string(hash),
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		log.Printf("create user error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	token, err := h.jwt.GenerateToken(user.ID)
	if err != nil {
		log.Printf("token error: %v", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.AuthResponse{
		Token: token,
		User: model.User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input model.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), input.Email)
	if err != nil {
		// Не говорим явно, что email не найден (безопасность)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	_ = h.userRepo.UpdateLastLogin(r.Context(), user.ID)

	token, err := h.jwt.GenerateToken(user.ID)
	if err != nil {
		log.Printf("token error: %v", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.AuthResponse{
		Token: token,
		User: model.User{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}
