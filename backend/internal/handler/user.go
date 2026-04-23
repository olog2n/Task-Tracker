package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"tracker/internal/repository"
	"tracker/internal/tracemiddleware"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// @Summary      Deactivate a user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string  "Admin only"
// @Failure      404  {object}  map[string]string
// @Router       /api/users/{id} [delete]
func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	currentUserID, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	currentUser, err := h.userRepo.GetByID(ctx, currentUserID)
	if err != nil {
		http.Error(w, "failed to get current user", http.StatusInternalServerError)
		return
	}
	if !currentUser.IsAdmin() {
		http.Error(w, "forbidden - admin only", http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if userID == currentUserID {
		http.Error(w, "cannot deactivate yourself", http.StatusBadRequest)
		return
	}

	_, err = h.userRepo.GetByID(ctx, userID)
	if err == sql.ErrNoRows {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.Deactivate(ctx, userID, currentUserID); err != nil {
		log.Printf("deactivation error: %v", err)
		http.Error(w, "failed to deactivate user", http.StatusInternalServerError)
		return
	}

	_ = h.userRepo.IncrementTokenVersion(ctx, userID)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Reactivate a user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/users/{id}/reactivate [post]
func (h *UserHandler) ReactivateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	currentUserID, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	currentUser, err := h.userRepo.GetByID(ctx, currentUserID)
	if err != nil {
		http.Error(w, "failed to get current user", http.StatusInternalServerError)
		return
	}
	if !currentUser.IsAdmin() {
		http.Error(w, "forbidden - admin only", http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if userID == currentUserID {
		http.Error(w, "cannot reactivate yourself", http.StatusBadRequest)
		return
	}

	targetUser, err := h.userRepo.GetByID(ctx, userID)
	if err == sql.ErrNoRows {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}
	if !targetUser.IsDeleted() {
		http.Error(w, "user is not deleted", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Reactivate(ctx, userID, currentUserID); err != nil {
		log.Printf("reactivation error: %v", err)
		http.Error(w, "failed to reactivate user", http.StatusInternalServerError)
		return
	}

	_ = h.userRepo.IncrementTokenVersion(ctx, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                     targetUser.ID,
		"email":                  targetUser.Email,
		"require_password_reset": true,
		"message":                "User reactivated. Password reset required on next login.",
	})
}
