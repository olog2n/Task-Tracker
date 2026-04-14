package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"tracker/internal/mocks"
	"tracker/internal/model"
)

func TestGetTasks_Success(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetAllFunc: func(ctx context.Context) ([]model.Task, error) {
			return []model.Task{
				{ID: 1, Title: "Test", Status: model.StatusBacklog},
			}, nil
		},
	}

	handler := NewTaskHandler(mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()

	handler.GetTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var tasks []model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Title != "Test" {
		t.Errorf("expected title 'Test', got '%s'", tasks[0].Title)
	}
}

func TestGetTasks_RepositoryError(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetAllFunc: func(ctx context.Context) ([]model.Task, error) {
			return nil, errors.New("database down")
		},
	}

	handler := NewTaskHandler(mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()

	handler.GetTasks(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var tasks []model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err == nil {
		t.Fatalf("expected to fail unmarshal: %v", err)
	}
}

func TestGetTasks_Empty(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetAllFunc: func(ctx context.Context) ([]model.Task, error) {
			return []model.Task{}, nil
		},
	}

	handler := NewTaskHandler(mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()

	handler.GetTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var tasks []model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("expected 0 task, got %d", len(tasks))
	}
}
