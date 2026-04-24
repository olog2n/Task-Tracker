package tracemiddleware

import (
	"context"
	"net/http"
	"strings"

	"tracker/internal/auth"
	"tracker/internal/model"

	"github.com/google/uuid"
)

func AuthMiddleware(jwt *auth.JWTService, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}

			if token == "" {
				cookie, err := r.Cookie(cookieName)
				if err == nil && cookie.Value != "" {
					token = cookie.Value
				}
			}

			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ValidateToken(token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserNameKey, claims.Name)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func GetUserEmailFromContext(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	return email, ok
}

func GetUserNameFromContext(r *http.Request) (string, bool) {
	name, ok := r.Context().Value(UserNameKey).(string)
	return name, ok
}

func GetUserFromContext(r *http.Request) (*model.User, bool) {
	userID, okID := GetUserIDFromContext(r)
	userEmail, okEmail := GetUserEmailFromContext(r)
	userName, okName := GetUserNameFromContext(r)

	if !okID || !okEmail || !okName {
		return nil, false
	}

	return &model.User{
		ID:    userID,
		Email: userEmail,
		Name:  userName,
	}, true
}
