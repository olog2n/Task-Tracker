package handler

import (
	"context"
	"net/http"
	"testing"
	"time"

	"tracker/internal/auth"
	"tracker/internal/tracemiddleware"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

const secret = "hWWDyfOjhNLBSIQ4vt4abLtDUZ9YAPA3aFotOl3m+r4="

func CreateTestToken(t *testing.T, userID int) string {
	t.Helper()
	claims := &auth.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	return tokenString
}

func CreateJWTService(t *testing.T) *auth.JWTService {
	t.Helper()
	jwtService, err := auth.NewJWTService(
		"HS256",
		secret,
		"", "", "",
		24*time.Hour,
		168*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}
	return jwtService
}

func WithURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func WithUserContext(req *http.Request, userID int) *http.Request {
	ctx := context.WithValue(req.Context(), tracemiddleware.UserIDKey, userID)
	return req.WithContext(ctx)
}
