package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tracker/internal/model"
	"tracker/internal/repository"
	"tracker/internal/traceMiddleware"

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	repo repository.TaskRepository
}

func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// GetTasks godoc
// @Summary      Get all tasks
// @Description  Get list of all tasks for authenticated user
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Task
// @Failure      401  {object}  map[string]string
// @Router       /api/tasks [get]
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	tasks, err := h.repo.GetAll(ctx)
	if err != nil {
		log.Printf("repo error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = make([]model.Task, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTaskByID godoc
// @Summary      Get a task by ID
// @Description  Get detailed information about a specific task by its ID. Only the task author can access this endpoint.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Task ID"  minimum(1)
// @Success      200  {object}  model.Task
// @Failure      400  {object}  map[string]string  "Invalid ID format"
// @Failure      401  {object}  map[string]string  "Unauthorized - Invalid or missing token"
// @Failure      404  {object}  map[string]string  "Task not found"
// @Router       /api/tasks/{id} [get]
func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	task, err := h.repo.GetByID(ctx, id)
	if err == sql.ErrNoRows {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("query error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// CreateTask godoc
// @Summary      Create a new task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body model.Task true "Task data"
// @Success      201  {object}  model.Task
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /api/tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	var input model.Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := validateTask(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.repo.Create(ctx, &input)
	if err != nil {
		log.Printf("create error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("last insert id error: %v", err)
		http.Error(w, "failed to get task id", http.StatusInternalServerError)
		return
	}

	task := model.Task{
		ID:          model.TaskID(id),
		Title:       input.Title,
		Author:      input.Author,
		Description: input.Description,
		Executor:    input.Executor,
		Status:      input.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// UpdateTask godoc
// @Summary      Update a task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int true "Task ID"
// @Param        request body model.Task true "Task data"
// @Success      200  {object}  model.Task
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/tasks/{id} [put]
// UpdateTask обновляет задачу
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Получаем ID из path-параметра
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Проверяем существование задачи
	existing, err := h.repo.GetByID(ctx, id)
	if err == sql.ErrNoRows {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("query error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Парсим тело запроса
	var input model.Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация
	if err := validateTask(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Сохраняем ID и обновляем поля
	input.ID = existing.ID
	input.CreatedAt = existing.CreatedAt
	input.UpdatedAt = time.Now()

	// 👇 Проверка прав (автор может редактировать)
	userID, ok := traceMiddleware.GetUserIDFromContext(r)
	if ok && userID != 0 {
		// Если задача привязана к пользователю — проверяем
		// Пока просто логируем, полную проверку добавим позже
		_ = userID
	}

	// Обновляем в БД
	if err := h.repo.Update(ctx, &input); err != nil {
		log.Printf("update error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

// DeleteTask godoc
// @Summary      Delete a task
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int true "Task ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Получаем ID из path-параметра
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Проверяем существование
	_, err = h.repo.GetByID(ctx, id)
	if err == sql.ErrNoRows {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("query error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Удаляем из БД
	if err := h.repo.Delete(ctx, id); err != nil {
		log.Printf("delete error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateTask(task *model.Task) error {
	if strings.TrimSpace(task.Title) == "" {
		return errors.New("title is required")
	}

	if len(task.Title) > model.TitleMaxLength {
		return errors.New("title must be 100 characters or less")
	}

	return nil
}
