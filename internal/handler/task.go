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

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	repo repository.TaskRepository
}

func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

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

func validateTask(task *model.Task) error {
	if strings.TrimSpace(task.Title) == "" {
		return errors.New("title is required")
	}

	if len(task.Title) > model.TitleMaxLength {
		return errors.New("title must be 100 characters or less")
	}

	return nil
}
