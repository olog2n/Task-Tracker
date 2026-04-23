package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tracker/internal/tracemiddleware"

	"github.com/golang-jwt/jwt/v5"
)

// CreateTestUser создаёт тестового пользователя
func CreateTestUser(email string, password string) map[string]interface{} {
	return map[string]interface{}{
		"email":    email,
		"password": password,
	}
}

// CreateTestProject создаёт тестовый проект
func CreateTestProject(name string, description string) map[string]interface{} {
	return map[string]interface{}{
		"name":        name,
		"description": description,
	}
}

// CreateTestTask создаёт тестовую задачу
func CreateTestTask(title string, description string, status string) map[string]interface{} {
	return map[string]interface{}{
		"title":       title,
		"description": description,
		"status":      status,
	}
}

// GenerateTestToken генерирует тестовый JWT токен
func GenerateTestToken(userID int, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// NewRequestWithAuth создаёт HTTP запрос с Authorization header
func NewRequestWithAuth(method, url string, body interface{}, token string) (*http.Request, error) {
	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// ExecuteRequest выполняет запрос и возвращает ответ
func ExecuteRequest(req *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// CheckResponseCode проверяет код ответа
func CheckResponseCode(t testing.TB, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// CheckResponseBody проверяет тело ответа
func CheckResponseBody(t testing.TB, expected, actual map[string]interface{}) {
	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)
	if string(expectedJSON) != string(actualJSON) {
		t.Errorf("Expected response body %s. Got %s\n", expectedJSON, actualJSON)
	}
}

// ContextWithUser создаёт контекст с user_id
func ContextWithUser(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, tracemiddleware.UserIDKey, userID)
}

// ContextWithProject создаёт контекст с project_id и member
func ContextWithProject(ctx context.Context, projectID int, member interface{}) context.Context {
	ctx = context.WithValue(ctx, tracemiddleware.ProjectIDKey, projectID)
	ctx = context.WithValue(ctx, tracemiddleware.ProjectMemberKey, member)
	return ctx
}
