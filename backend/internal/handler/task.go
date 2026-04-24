package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"tracker/internal/audit"
	"tracker/internal/model"
	"tracker/internal/repository"
	"tracker/internal/tracemiddleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TaskHandler struct {
	repo        repository.TaskRepository
	auditLogger *audit.Logger
}

func NewTaskHandler(
	repo repository.TaskRepository,
	auditLogger *audit.Logger,
) *TaskHandler {
	return &TaskHandler{
		repo:        repo,
		auditLogger: auditLogger,
	}
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
// func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	ctx := r.Context()

// 	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
// 	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

// 	if limit <= 0 {
// 		limit = 20
// 	}
// 	if limit > 100 {
// 		limit = 100
// 	}
// 	if offset < 0 {
// 		offset = 0
// 	}

// 	tasks, err := h.repo.GetWithPagination(ctx, limit, offset)
// 	if err != nil {
// 		log.Printf("repo error: %v", err)
// 		http.Error(w, "database error", http.StatusInternalServerError)
// 		return
// 	}

// 	total, _ := h.repo.Count(ctx)

// 	if tasks == nil {
// 		tasks = make([]model.Task, 0)
// 	}

// 	response := model.PaginatedTasks{
// 		Tasks:  tasks,
// 		Total:  total,
// 		Limit:  limit,
// 		Offset: offset,
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }

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
	ctx := r.Context()
	user, _ := tracemiddleware.GetUserFromContext(r)

	id, err := parseUUIDParam(r, "id")
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.repo.GetByID(ctx, id)
	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		log.Printf("query error: %v", err)
		RespondError(w, http.StatusInternalServerError, "database error")
		return
	}

	h.auditLogger.LogTaskViewed(ctx, r, user, task)
	RespondJSON(w, http.StatusOK, task)
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
// ============================================================================
// UpdateTask — обновление задачи
// ============================================================================
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Получаем userID из контекста (через middleware)
	userID, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok || userID == uuid.Nil {
		RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Парсим ID задачи из URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	// Получаем существующую задачу (через repo)
	existing, err := h.repo.GetByID(ctx, id)
	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		log.Printf("query error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Проверка прав: только создатель может редактировать
	//TODO: Изменить поведение, нужно чтобы изменять задачу мог либо автор, либо кто-то выше него, например админ
	if existing.CreatedBy != userID {
		RespondError(w, http.StatusForbidden, "forbidden - only author can update this task")
		return
	}

	// Парсим входные данные
	var input model.TaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Валидация
	if err := validateTaskInput(&input); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedTask, err := h.repo.Update(ctx, id, &input, userID)
	if err != nil {
		log.Printf("update error: %v", err)
		RespondError(w, http.StatusInternalServerError, "database error")
		return
	}

	user, ok := tracemiddleware.GetUserFromContext(r)
	if !ok || user.ID == uuid.Nil {
		RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	h.auditLogger.LogTaskUpdated(ctx, r, user, existing, updatedTask)

	// Если статус изменился — отдельный лог
	if existing.StatusID != updatedTask.StatusID {
		h.auditLogger.LogStatusChanged(ctx, r, user, id, existing.StatusID, updatedTask.StatusID, "")
	}

	RespondJSON(w, http.StatusOK, updatedTask)
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
	ctx := r.Context()

	user, ok := tracemiddleware.GetUserFromContext(r)
	if !ok || user.ID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid task id")
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

	// Проверка прав: только создатель может удалять задачу
	// TODO: Добавить проверку на Project Owner / Admin
	if task.CreatedBy != user.ID {
		http.Error(w, "forbidden - only author can delete this task", http.StatusForbidden)
		return
	}

	if err := h.repo.Delete(ctx, id, user.ID); err != nil {
		log.Printf("delete error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	h.auditLogger.LogTaskDeleted(ctx, r, user, task)

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// GetAll — получение списка задач
// ============================================================================
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := tracemiddleware.GetUserFromContext(r)

	filter := &model.TaskFilter{Limit: 50, Offset: 0}

	if projectID, err := parseUUIDQuery(r, "project_id"); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid project_id")
		return
	} else if projectID != nil {
		filter.ProjectID = projectID
	}

	if assigneeID := r.URL.Query().Get("assignee_id"); assigneeID != "" {
		id, err := strconv.Atoi(assigneeID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid assignee_id")
			return
		}
		filter.AssigneeID = &id
	}

	if statusID, err := parseUUIDQuery(r, "status_id"); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid status_id")
		return
	} else if statusID != nil {
		filter.StatusID = statusID
	}

	filter.Priority = r.URL.Query().Get("priority")
	filter.Search = r.URL.Query().Get("search")
	filter.SortBy = r.URL.Query().Get("sort_by")
	filter.SortOrder = r.URL.Query().Get("sort_order")

	if limit := r.URL.Query().Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil || l < 1 || l > 200 {
			http.Error(w, "limit must be between 1 and 200", http.StatusBadRequest)
			return
		}
		filter.Limit = l
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil || o < 0 {
			http.Error(w, "offset must be non-negative", http.StatusBadRequest)
			return
		}
		filter.Offset = o
	}

	paginatedTasks, err := h.repo.GetAll(ctx, filter)
	if err != nil {
		log.Printf("failed to get tasks: %v", err)
		RespondError(w, http.StatusInternalServerError, "failed to get tasks")
		return
	}

	h.auditLogger.LogBulkSelect(ctx, r, user, "task", paginatedTasks.Total, map[string]interface{}{
		"project_id":  filter.ProjectID,
		"assignee_id": filter.AssigneeID,
		"status_id":   filter.StatusID,
		"priority":    filter.Priority,
		"has_search":  filter.Search != "",
	})

	RespondJSON(w, http.StatusOK, paginatedTasks)
}

// ============================================================================
// Create — создание задачи
// ============================================================================

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
	ctx := r.Context()
	user, _ := tracemiddleware.GetUserFromContext(r)

	var input model.TaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateTaskInput(&input); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	projectID := uuid.Nil
	if pid, err := parseUUIDQuery(r, "project_id"); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid project_id")
		return
	} else if pid != nil {
		projectID = *pid
	} else {
		// TODO: получить дефолтный проект из конфига/контекста
		// Для демо: используем фиксированный UUID
		projectID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	statusID := uuid.Nil
	if input.StatusID != nil {
		statusID = *input.StatusID
	} else {
		// TODO: получить дефолтный статус из процесса проекта
		statusID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	task := &model.Task{
		Title:          input.Title,
		Description:    input.Description,
		StatusID:       statusID,
		Priority:       input.Priority,
		ProjectID:      projectID,
		AssigneeID:     input.AssigneeID,
		CreatedBy:      user.ID,
		Classification: model.ClassificationInternal,
		IsSensitive:    false,
	}

	if err := h.repo.Create(ctx, task); err != nil {
		log.Printf("failed to create task: %v", err)
		RespondError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	h.auditLogger.LogTaskCreated(ctx, r, user, task)
	RespondJSON(w, http.StatusCreated, task)
}

func validateTaskInput(input *model.TaskInput) error {
	// Title: обязательный, 1-500 символов
	if input.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(input.Title) > 500 {
		return fmt.Errorf("title must be less than 500 characters")
	}

	// Description: опциональный, макс. 10000 символов
	if len(input.Description) > 10000 {
		return fmt.Errorf("description must be less than 10000 characters")
	}

	// Priority: опциональный, но если есть — только допустимые значения
	if input.Priority != "" {
		switch input.Priority {
		case "low", "medium", "high":
			// OK
		default:
			return fmt.Errorf("priority must be one of: low, medium, high")
		}
	}

	return nil
}

// parseUUIDParam — парсит UUID из chi URL param
func parseUUIDParam(r *http.Request, paramName string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, paramName)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s: %w", paramName, err)
	}
	return id, nil
}

// parseUUIDQuery — парсит UUID из query param
func parseUUIDQuery(r *http.Request, paramName string) (*uuid.UUID, error) {
	val := r.URL.Query().Get(paramName)
	if val == "" {
		return nil, nil
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", paramName, err)
	}
	return &id, nil
}
