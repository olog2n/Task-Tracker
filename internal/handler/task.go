package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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

func (h TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tasks, err := h.repo.GetAll(ctx)
	if err != nil {
		log.Printf("query error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("failed to encode: %v", err)
		return
	}
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
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

	task, err := h.repo.GetById(ctx, id)

	if err != nil {
		log.Printf("query error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if task == nil {
		http.Error(w, "task not found", 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input model.Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	result, err := h.repo.Create(ctx, &input)

	if err != nil {
		log.Printf("insert error: %v", err)
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
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}
