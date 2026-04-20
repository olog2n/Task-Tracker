package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"tracker/internal/auth"
	"tracker/internal/model"
	"tracker/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo          repository.UserRepository
	jwt               *auth.JWTService
	cookieSecure      bool
	cookieDomain      string
	cookiePath        string
	cookieNameAccess  string
	cookieNameRefresh string
}

func NewAuthHandler(
	userRepo repository.UserRepository,
	jwt *auth.JWTService,
	cookieSecure bool,
	cookieDomain string,
	cookiePath string,
	cookieNameAccess string,
	cookieNameRefresh string,
) *AuthHandler {
	if cookiePath == "" {
		cookiePath = "/"
	}
	if cookieNameAccess == "" {
		cookieNameAccess = "access_token"
	}
	if cookieNameRefresh == "" {
		cookieNameRefresh = "refresh_token"
	}

	return &AuthHandler{
		userRepo:          userRepo,
		jwt:               jwt,
		cookieSecure:      cookieSecure,
		cookieDomain:      cookieDomain,
		cookiePath:        cookiePath,
		cookieNameAccess:  cookieNameAccess,
		cookieNameRefresh: cookieNameRefresh,
	}
}

func (h *AuthHandler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  expiresAt,
		Path:     "/api",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
		Domain:   h.cookieDomain,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(REFRESH_TOKEN_EXPIRE),
		Path:     "/auth/refresh",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
		Domain:   h.cookieDomain,
	})
}

func (h *AuthHandler) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieNameAccess,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     h.cookiePath,
		Domain:   h.cookieDomain,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieNameRefresh,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/auth/refresh",
		Domain:   h.cookieDomain,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.RegisterInput true "Registration data"
// @Success      201  {object}  model.AuthResponse
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /api/auth/register [post]
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

	tokenPair, err := h.jwt.GenerateTokenPair(user.ID)
	if err != nil {
		log.Printf("token error: %v", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	h.setAuthCookies(w, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresAt)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// Login godoc
// @Summary      Login user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.LoginInput true "Login credentials"
// @Success      200  {object}  model.AuthResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /api/auth/login [post]
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
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.CanLogin() {
		http.Error(w, "account deactivated", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	_ = h.userRepo.UpdateLastLogin(r.Context(), user.ID)

	tokenPair, err := h.jwt.GenerateTokenPair(user.ID)
	if err != nil {
		log.Printf("token error: %v", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	h.setAuthCookies(w, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresAt)

	_ = h.userRepo.UpdateLastLogin(r.Context(), user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.User{
		ID:    user.ID,
		Email: user.Email,
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  auth.TokenPair
// @Failure      401  {object}  map[string]string
// @Router       /api/auth/refresh [post]
// RefreshToken обновляет пару токенов
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "refresh token required", http.StatusUnauthorized)
		return
	}

	pair, err := h.jwt.RefreshAccessToken(refreshCookie.Value)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	h.setAuthCookies(w, pair.AccessToken, pair.RefreshToken, pair.ExpiresAt)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]time.Time{
		"expires_at": pair.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearAuthCookies(w)

	// h.tokenBlacklist.Add(refreshToken)

	w.WriteHeader(http.StatusNoContent)
}
