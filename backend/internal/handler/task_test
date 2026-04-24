package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tracker/internal/mocks"
	"tracker/internal/model"
)

func TestCreateTask_Validation_EmptyTitle(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{}
	taskHandler := NewTaskHandler(mockRepo)

	input := model.Task{Title: ""}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.CreateTask(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestCreateTask_Validation_TitleTooLong(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{}
	taskHandler := NewTaskHandler(mockRepo)

	longTitle := strings.Repeat("a", 101)
	input := model.Task{Title: longTitle}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.CreateTask(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestCreateTask_Validation_Valid(t *testing.T) {
	var capturedTask *model.Task
	mockRepo := &mocks.MockTaskRepository{
		CreateFunc: func(ctx context.Context, task *model.Task) (sql.Result, error) {
			capturedTask = task
			return &mockResult{id: 1}, nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	input := model.Task{
		Title:       "Valid Task",
		Description: "Test description",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.CreateTask(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	if capturedTask == nil {
		t.Fatal("Expected Create to be called")
	}
}

func TestGetTasks_Success(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetWithPaginationFunc: func(ctx context.Context, limit, offset int) ([]model.Task, error) {
			return []model.Task{
				{ID: 1, Title: "Task 1", Author: "user@example.com", Status: model.StatusBacklog},
				{ID: 2, Title: "Task 2", Author: "user@example.com", Status: model.StatusInProgress},
			}, nil
		},
		CountFunc: func(ctx context.Context) (int, error) {
			return 2, nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.GetTasks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response model.PaginatedTasks
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(response.Tasks))
	}
	if response.Total != 2 {
		t.Errorf("Expected total 2, got %d", response.Total)
	}
}

func TestGetTasks_RepositoryError(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetWithPaginationFunc: func(ctx context.Context, limit, offset int) ([]model.Task, error) {
			return nil, sql.ErrConnDone
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.GetTasks(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
}

func TestGetTasks_EmptyList(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetWithPaginationFunc: func(ctx context.Context, limit, offset int) ([]model.Task, error) {
			return []model.Task{}, nil
		},
		CountFunc: func(ctx context.Context) (int, error) {
			return 0, nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.GetTasks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response model.PaginatedTasks
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(response.Tasks))
	}
	if response.Total != 0 {
		t.Errorf("Expected total 0, got %d", response.Total)
	}
}

func TestGetTasks_Pagination(t *testing.T) {
	var capturedLimit, capturedOffset int

	mockRepo := &mocks.MockTaskRepository{
		GetWithPaginationFunc: func(ctx context.Context, limit, offset int) ([]model.Task, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []model.Task{}, nil
		},
		CountFunc: func(ctx context.Context) (int, error) {
			return 0, nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks?limit=10&offset=20", nil)
	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.GetTasks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if capturedLimit != 10 {
		t.Errorf("Expected limit 10, got %d", capturedLimit)
	}
	if capturedOffset != 20 {
		t.Errorf("Expected offset 20, got %d", capturedOffset)
	}
}

func TestGetTaskByID_Success(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return &model.Task{
				ID:          1,
				Title:       "Test Task",
				Author:      "user@example.com",
				Description: "Test description",
				Status:      model.StatusBacklog,
			}, nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1", nil)
	req = WithUserContext(req, 1)
	req = WithURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	taskHandler.GetTaskByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var task model.Task
	if err := json.Unmarshal(rr.Body.Bytes(), &task); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", task.Title)
	}
}

func TestGetTaskByID_NotFound(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return nil, sql.ErrNoRows
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/999", nil)
	req = WithUserContext(req, 1)
	req = WithURLParam(req, "id", "999")

	rr := httptest.NewRecorder()
	taskHandler.GetTaskByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestGetTaskByID_RepositoryError(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return nil, sql.ErrConnDone
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1", nil)
	req = WithUserContext(req, 1)
	req = WithURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	taskHandler.GetTaskByID(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
}

func TestGetTaskByID_InvalidID(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/abc", nil)
	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.GetTaskByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestUpdateTask_Success(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return &model.Task{
				ID:       1,
				Title:    "Old Title",
				AuthorID: sql.NullInt64{Int64: 1, Valid: true}, // Тот же пользователь
			}, nil
		},
		UpdateFunc: func(ctx context.Context, task *model.Task) error {
			return nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	input := model.Task{Title: "New Title", Description: "New description"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = WithUserContext(req, 1)
	req = WithURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	taskHandler.UpdateTask(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestUpdateTask_Forbidden(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return &model.Task{
				ID:       1,
				Title:    "Old Title",
				AuthorID: sql.NullInt64{Int64: 999, Valid: true}, // Другой пользователь
			}, nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	input := model.Task{Title: "New Title"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = WithUserContext(req, 1) // user_id = 1, но author_id = 999
	req = WithURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	taskHandler.UpdateTask(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusForbidden, rr.Code, rr.Body.String())
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return nil, sql.ErrNoRows
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	input := model.Task{Title: "New Title"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = WithUserContext(req, 1)
	req = WithURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	taskHandler.UpdateTask(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestDeleteTask_Success(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return &model.Task{
				ID:       1,
				Title:    "Task to delete",
				AuthorID: sql.NullInt64{Int64: 1, Valid: true}, // Тот же пользователь
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/1", nil)
	req = WithUserContext(req, 1)
	req = WithURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	taskHandler.DeleteTask(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNoContent, rr.Code, rr.Body.String())
	}
}

func TestDeleteTask_Forbidden(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return &model.Task{
				ID:       1,
				Title:    "Task to delete",
				AuthorID: sql.NullInt64{Int64: 999, Valid: true}, // Другой пользователь
			}, nil
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/1", nil)
	req = WithUserContext(req, 1) // user_id = 1, но author_id = 999
	req = WithURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	taskHandler.DeleteTask(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusForbidden, rr.Code, rr.Body.String())
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*model.Task, error) {
			return nil, sql.ErrNoRows
		},
	}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/999", nil)
	req = WithUserContext(req, 1)
	req = WithURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	taskHandler.DeleteTask(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestCreateTask_MethodNotAllowed(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.CreateTask(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestGetTasks_MethodNotAllowed(t *testing.T) {
	mockRepo := &mocks.MockTaskRepository{}
	taskHandler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", nil)
	req = WithUserContext(req, 1)

	rr := httptest.NewRecorder()
	taskHandler.GetTasks(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

type mockResult struct {
	id int64
}

func (m *mockResult) LastInsertId() (int64, error) {
	return m.id, nil
}

func (m *mockResult) RowsAffected() (int64, error) {
	return 1, nil
}
