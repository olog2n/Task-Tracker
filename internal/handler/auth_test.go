package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tracker/internal/auth"

	"github.com/golang-jwt/jwt/v5"

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

func TestAuthHandler_Refresh_Success(t *testing.T) {
	userID := 1

	mockRepo := &mocks.MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.User, error) {
			return &model.User{
				ID:       userID,
				Email:    "test@example.com",
				IsActive: true,
			}, nil
		},
	}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")
	//TODO: REFACTORING, MUST BE IN UNIFIED FORM
	tokenPair, err := jwtService.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: tokenPair.RefreshToken,
		Path:  "/",
	})

	rr := httptest.NewRecorder()
	authHandler.RefreshToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	cookies := rr.Result().Cookies()
	var hasAccessToken, hasRefreshToken bool
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			hasAccessToken = true
		}
		if cookie.Name == "refresh_token" {
			hasRefreshToken = true
		}
	}

	if !hasAccessToken {
		t.Error("Expected access_token cookie in response")
	}
	if !hasRefreshToken {
		t.Error("Expected refresh_token cookie in response")
	}
}

func TestAuthHandler_Refresh_MissingCookie(t *testing.T) {
	mockRepo := &mocks.MockUserRepository{}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)

	rr := httptest.NewRecorder()
	authHandler.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	mockRepo := &mocks.MockUserRepository{}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "invalid-token",
		Path:  "/",
	})

	rr := httptest.NewRecorder()
	authHandler.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Refresh_ExpiredToken(t *testing.T) {
	mockRepo := &mocks.MockUserRepository{}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	// Создаём просроченный токен вручную
	claims := &auth.Claims{
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString([]byte("test-secret-key-min-32-characters-long!"))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: expiredToken,
		Path:  "/",
	})

	rr := httptest.NewRecorder()
	authHandler.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Refresh_TokenReuse(t *testing.T) {
	//TODO: Check this test after add external cache (Redis/Memory)
	t.Skip("Requires blacklist integration - test in integration suite")
	userID := 1

	mockRepo := &mocks.MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.User, error) {
			return &model.User{
				ID:       userID,
				Email:    "test@example.com",
				IsActive: true,
			}, nil
		},
	}

	jwtService := CreateJWTService(t)
	authHandler := NewAuthHandler(mockRepo, jwtService, false, "", "/", "access_token", "refresh_token")

	tokenPair, err := jwtService.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req1.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: tokenPair.RefreshToken,
		Path:  "/",
	})
	rr1 := httptest.NewRecorder()
	authHandler.RefreshToken(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First refresh: Expected status %d, got %d", http.StatusOK, rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req2.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: tokenPair.RefreshToken,
		Path:  "/",
	})
	rr2 := httptest.NewRecorder()
	authHandler.RefreshToken(rr2, req2)

	if rr2.Code != http.StatusUnauthorized {
		t.Errorf("Second refresh: Expected status %d, got %d. Body: %s", http.StatusUnauthorized, rr2.Code, rr2.Body.String())
	}
}
