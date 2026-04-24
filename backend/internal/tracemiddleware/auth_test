package tracemiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tracker/internal/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "hWWDyfOjhNLBSIQ4vt4abLtDUZ9YAPA3aFotOl3m+r4="
	jwtService, err := auth.NewJWTService(
		"HS256",
		secret,
		"",
		"",
		"",
		24*time.Hour,
		168*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	claims := auth.Claims{
		UserID: 123,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserIDFromContext(r)
		assert.True(t, ok)
		assert.Equal(t, 123, userID)
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(jwtService, "access_token")(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	secret := "hWWDyfOjhNLBSIQ4vt4abLtDUZ9YAPA3aFotOl3m+r4="
	jwtService, err := auth.NewJWTService(
		"HS256",
		secret,
		"",
		"",
		"",
		24*time.Hour,
		168*time.Hour,
	)
	if err != nil {
		t.Fatalf("jwtservice error %v", err)
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(jwtService, "access_token")(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	secret := "hWWDyfOjhNLBSIQ4vt4abLtDUZ9YAPA3aFotOl3m+r4="
	jwtService, err := auth.NewJWTService("HS256", secret, "", "", "", 24*time.Hour, 168*time.Hour)
	if err != nil {
		t.Fatalf("jwtservice error %v", err)
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(jwtService, "access_token")(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	secret := "hWWDyfOjhNLBSIQ4vt4abLtDUZ9YAPA3aFotOl3m+r4="
	jwtService, err := auth.NewJWTService("HS256", secret, "", "", "", 24*time.Hour, 168*time.Hour)
	if err != nil {
		t.Fatalf("jwtservice error %v", err)
	}

	claims := auth.Claims{
		UserID: 123,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Уже истёк
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(jwtService, "access_token")(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
