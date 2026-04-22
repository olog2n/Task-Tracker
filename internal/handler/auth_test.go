package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tracker/internal/auth"
	"tracker/internal/mocks"
	"tracker/internal/model"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	mockRepo := &mocks.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return nil, sql.ErrNoRows // Пользователя нет
		},
		CreateFunc: func(ctx context.Context, user *model.User) error {
			return nil
		},
	}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	input := model.RegisterInput{Email: "test@example.com", Password: "password123"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authHandler.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Register_EmailExists(t *testing.T) {
	mockRepo := &mocks.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {

			t.Logf("GetByEmail called with email: %s", email)

			return &model.User{
				ID:       1,
				Email:    email,
				IsActive: true,
			}, nil
		},
	}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	input := model.RegisterInput{Email: "existing@example.com", Password: "password123"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authHandler.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusConflict, rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	hashedPassword, _ := auth.HashPassword("password123")

	mockRepo := &mocks.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashedPassword,
				IsActive:     true,
				TokenVersion: 1,
			}, nil
		},
		UpdateLastLoginFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	input := model.LoginInput{Email: "test@example.com", Password: "password123"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authHandler.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response model.AuthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.AccessToken == "" {
		t.Error("Expected access token in response")
	}
	if response.RefreshToken == "" {
		t.Error("Expected refresh token in response")
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	hashedPassword, _ := auth.HashPassword("correctpassword")

	mockRepo := &mocks.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashedPassword,
				IsActive:     true,
			}, nil
		},
	}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	input := model.LoginInput{Email: "test@example.com", Password: "wrongpassword"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authHandler.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Login_DeactivatedAccount(t *testing.T) {
	hashedPassword, _ := auth.HashPassword("password123")

	mockRepo := &mocks.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashedPassword,
				IsActive:     false,
			}, nil
		},
	}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	input := model.LoginInput{Email: "test@example.com", Password: "password123"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authHandler.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}
