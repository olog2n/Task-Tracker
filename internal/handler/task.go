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
	"tracker/internal/tracemiddleware"

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	repo repository.TaskRepository
}

type PaginatedTasks struct {
	Tasks  []model.Task `json:"tasks"`
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// GetTasks godoc
// @Summary      Get all tasks
// @Description  Get list of all tasks with pagination
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query    int  false  "Items per page"  default(20)  minimum(1)  maximum(100)
// @Param        offset query    int  false  "Offset"          default(0)   minimum(0)
// @Success      200  {object}  PaginatedTasks
// @Failure      401  {object}  map[string]string
// @Router       /api/tasks [get]
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	tasks, err := h.repo.GetAll(ctx)
	if err != nil {
		log.Printf("repo error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	total, _ := h.repo.Count(ctx)

	if tasks == nil {
		tasks = make([]model.Task, 0)
	}

	response := PaginatedTasks{
		Tasks:  tasks,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	userID, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input model.Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := validateTask(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.AuthorID = userID
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()
	input.Status = model.StatusBacklog

	result, err := h.repo.Create(ctx, &input)
	if err != nil {
		log.Printf("create error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	input.ID = model.TaskID(id)
	if err != nil {
		log.Printf("last insert id error: %v", err)
		http.Error(w, "failed to get task id", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
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
// @Failure      403  {object}  map[string]string  "Forbidden - Not the author"
// @Failure      404  {object}  map[string]string
// @Router       /api/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	userID, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

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

	if existing.AuthorID != userID {
		http.Error(w, "forbidden - only author can update this task", http.StatusForbidden)
		return
	}

	var input model.Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := validateTask(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.ID = existing.ID
	input.AuthorID = existing.AuthorID
	input.CreatedAt = existing.CreatedAt
	input.UpdatedAt = time.Now()

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
// @Failure      403  {object}  map[string]string  "Forbidden - Not the author"
// @Failure      404  {object}  map[string]string
// @Router       /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	userID, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
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

	if task.AuthorID != userID {
		http.Error(w, "forbidden - only author can delete this task", http.StatusForbidden)
		return
	}

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
