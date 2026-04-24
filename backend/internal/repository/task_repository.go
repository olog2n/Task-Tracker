package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"tracker/internal/model"

	"github.com/google/uuid"
)

type TaskRepository interface {
	GetWithPagination(ctx context.Context, limit, offset int) ([]*model.Task, error)
	Count(ctx context.Context) (int, error)

	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	GetAll(ctx context.Context, filter *model.TaskFilter) (*model.PaginatedTasks, error)
	Update(ctx context.Context, id uuid.UUID, input *model.TaskInput, updatedBy uuid.UUID) (*model.Task, error)
	Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	GetByProject(ctx context.Context, projectID uuid.UUID) ([]*model.Task, error)
	GetByAssignee(ctx context.Context, assigneeID uuid.UUID) ([]*model.Task, error)

	GetAvailableStatuses(ctx context.Context, projectID uuid.UUID) ([]*model.Status, error)
	GetTransitionsForStatus(ctx context.Context, projectID, statusID uuid.UUID) ([]*model.Transition, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

// ============================================================================
// GetWithPagination — СТАРЫЙ метод (сохраняем для совместимости)
// ============================================================================
func (r *taskRepository) GetWithPagination(ctx context.Context, limit, offset int) ([]*model.Task, error) {
	query := `
		SELECT 
			t.id, t.title, t.description, t.status_id, t.priority,
			t.project_id, t.assignee_id, t.created_by, t.updated_by,
			t.classification, t.is_sensitive, t.created_at, t.updated_at,
			s.id as status_id, s.name as status_name, s.color as status_color
		FROM tasks t
		LEFT JOIN statuses s ON t.status_id = s.id
		WHERE t.deleted_at IS NULL
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	return r.scanTasksWithStatus(rows)
}

// ============================================================================
// Count — СТАРЫЙ метод (сохраняем для совместимости)
// ============================================================================
func (r *taskRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM tasks WHERE deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// ============================================================================
// Create — создание задачи
// ============================================================================
func (r *taskRepository) Create(ctx context.Context, task *model.Task) error {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}

	query := `
		INSERT INTO tasks 
		(id, title, description, status_id, priority, project_id, assignee_id, 
		 created_by, classification, is_sensitive, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		task.ID.String(), task.Title, task.Description, task.StatusID.String(),
		task.Priority, task.ProjectID.String(), task.AssigneeID, task.CreatedBy,
		task.Classification, task.IsSensitive, now, now)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	task.CreatedAt = now
	task.UpdatedAt = now

	return nil
}

// ============================================================================
// GetByID — получение задачи по ID (с развёрнутым статусом)
// ============================================================================
func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	query := `
		SELECT 
			t.id, t.title, t.description, t.status_id, t.priority,
			t.project_id, t.assignee_id, t.created_by, t.updated_by,
			t.classification, t.is_sensitive, t.created_at, t.updated_at,
			s.id as status_id, s.name as status_name, s.color as status_color,
			s.order as status_order, s.is_final as status_is_final
		FROM tasks t
		LEFT JOIN statuses s ON t.status_id = s.id
		WHERE t.id = ? AND t.deleted_at IS NULL
	`

	task := &model.Task{}
	var statusID uuid.UUID
	var statusName, statusColor sql.NullString
	var statusOrder int
	var statusIsFinal sql.NullBool

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.StatusID, &task.Priority,
		&task.ProjectID, &task.AssigneeID, &task.CreatedBy, &task.UpdatedBy,
		&task.Classification, &task.IsSensitive, &task.CreatedAt, &task.UpdatedAt,
		&statusID, &statusName, &statusColor, &statusOrder, &statusIsFinal)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Разворачиваем статус если есть
	if statusName.Valid {
		task.Status = &model.Status{
			ID:      statusID,
			Name:    statusName.String,
			Color:   statusColor.String,
			Order:   statusOrder,
			IsFinal: statusIsFinal.Valid && statusIsFinal.Bool,
		}
	}

	return task, nil
}

// ============================================================================
// GetAll — получение списка задач с фильтрами и пагинацией (НОВЫЙ метод v2)
// ============================================================================
func (r *taskRepository) GetAll(ctx context.Context, filter *model.TaskFilter) (*model.PaginatedTasks, error) {
	if filter == nil {
		filter = &model.TaskFilter{
			Limit:  50,
			Offset: 0,
		}
	}

	// Базовый запрос
	query := `
		SELECT 
			t.id, t.title, t.description, t.status_id, t.priority,
			t.project_id, t.assignee_id, t.created_by, t.updated_by,
			t.classification, t.is_sensitive, t.created_at, t.updated_at,
			s.id as status_id, s.name as status_name, s.color as status_color
		FROM tasks t
		LEFT JOIN statuses s ON t.status_id = s.id
		WHERE t.deleted_at IS NULL
	`

	countQuery := `SELECT COUNT(*) FROM tasks t WHERE t.deleted_at IS NULL`

	args := []interface{}{}
	countArgs := []interface{}{}

	// Динамические фильтры
	if filter.ProjectID != nil {
		query += " AND t.project_id = ?"
		countQuery += " AND t.project_id = ?"
		args = append(args, *filter.ProjectID)
		countArgs = append(countArgs, *filter.ProjectID)
	}

	if filter.AssigneeID != nil {
		query += " AND t.assignee_id = ?"
		countQuery += " AND t.assignee_id = ?"
		args = append(args, *filter.AssigneeID)
		countArgs = append(countArgs, *filter.AssigneeID)
	}

	if filter.StatusID != nil {
		query += " AND t.status_id = ?"
		countQuery += " AND t.status_id = ?"
		args = append(args, *filter.StatusID)
		countArgs = append(countArgs, *filter.StatusID)
	}

	if filter.Priority != "" {
		query += " AND t.priority = ?"
		countQuery += " AND t.priority = ?"
		args = append(args, filter.Priority)
		countArgs = append(countArgs, filter.Priority)
	}

	if filter.Search != "" {
		query += " AND (t.title LIKE ? OR t.description LIKE ?)"
		countQuery += " AND (t.title LIKE ? OR t.description LIKE ?)"
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}

	if filter.Classification != "" {
		query += " AND t.classification = ?"
		countQuery += " AND t.classification = ?"
		args = append(args, filter.Classification)
		countArgs = append(countArgs, filter.Classification)
	}

	if filter.Sensitive != nil {
		query += " AND t.is_sensitive = ?"
		countQuery += " AND t.is_sensitive = ?"
		args = append(args, *filter.Sensitive)
		countArgs = append(countArgs, *filter.Sensitive)
	}

	// Сортировка
	switch filter.SortBy {
	case "created_at", "updated_at", "priority", "title":
		query += fmt.Sprintf(" ORDER BY t.%s", filter.SortBy)
		if filter.SortOrder == "DESC" {
			query += " DESC"
		} else {
			query += " ASC"
		}
	default:
		query += " ORDER BY t.created_at DESC"
	}

	// Пагинация
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	tasks, err := r.scanTasksWithStatus(rows)
	if err != nil {
		return nil, err
	}

	return &model.PaginatedTasks{
		Tasks:   tasks,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: filter.Offset+len(tasks) < total,
	}, nil
}

// ============================================================================
// Update — обновление задачи
// ============================================================================
func (r *taskRepository) Update(ctx context.Context, id uuid.UUID, input *model.TaskInput, updatedBy uuid.UUID) (*model.Task, error) {
	updates := []string{}
	args := []interface{}{}

	if input.Title != "" {
		updates = append(updates, "title = ?")
		args = append(args, input.Title)
	}
	if input.Description != "" {
		updates = append(updates, "description = ?")
		args = append(args, input.Description)
	}
	if input.StatusID != nil {
		updates = append(updates, "status_id = ?")
		args = append(args, input.StatusID.String())
	}
	if input.Priority != "" {
		updates = append(updates, "priority = ?")
		args = append(args, input.Priority)
	}
	if input.AssigneeID != nil {
		updates = append(updates, "assignee_id = ?")
		args = append(args, *input.AssigneeID)
	}

	updates = append(updates, "updated_at = ?", "updated_by = ?")
	args = append(args, time.Now(), updatedBy)

	if len(updates) < 2 {
		return r.GetByID(ctx, id)
	}

	query := fmt.Sprintf(`UPDATE tasks SET %s WHERE id = ?`, strings.Join(updates, ", "))
	args = append(args, id.String())

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return r.GetByID(ctx, id)
}

// ============================================================================
// Delete — мягкое удаление задачи (soft delete)
// ============================================================================
func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	query := `UPDATE tasks SET deleted_at = ?, deleted_by = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, time.Now(), deletedBy, id)
	return err
}

// ============================================================================
// GetByProject — получение задач проекта
// ============================================================================
func (r *taskRepository) GetByProject(ctx context.Context, projectID uuid.UUID) ([]*model.Task, error) {
	query := `
		SELECT 
			t.id, t.title, t.description, t.status_id, t.priority,
			t.project_id, t.assignee_id, t.created_by, t.updated_by,
			t.classification, t.is_sensitive, t.created_at, t.updated_at,
			s.id as status_id, s.name as status_name, s.color as status_color
		FROM tasks t
		LEFT JOIN statuses s ON t.status_id = s.id
		WHERE t.project_id = ? AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTasksWithStatus(rows)
}

// ============================================================================
// GetByAssignee — получение задач исполнителя
// ============================================================================
func (r *taskRepository) GetByAssignee(ctx context.Context, assigneeID uuid.UUID) ([]*model.Task, error) {
	query := `
		SELECT 
			t.id, t.title, t.description, t.status_id, t.priority,
			t.project_id, t.assignee_id, t.created_by, t.updated_by,
			t.classification, t.is_sensitive, t.created_at, t.updated_at,
			s.id as status_id, s.name as status_name, s.color as status_color
		FROM tasks t
		LEFT JOIN statuses s ON t.status_id = s.id
		WHERE t.assignee_id = ? AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, assigneeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTasksWithStatus(rows)
}

// ============================================================================
// Helpers: scanTasksWithStatus — сканирует rows с развёрнутым статусом
// ============================================================================
func (r *taskRepository) scanTasksWithStatus(rows *sql.Rows) ([]*model.Task, error) {
	var tasks []*model.Task

	for rows.Next() {
		task := &model.Task{}
		var statusID uuid.UUID
		var statusName, statusColor sql.NullString

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.StatusID, &task.Priority,
			&task.ProjectID, &task.AssigneeID, &task.CreatedBy, &task.UpdatedBy,
			&task.Classification, &task.IsSensitive, &task.CreatedAt, &task.UpdatedAt,
			&statusID, &statusName, &statusColor)

		if err != nil {
			return nil, err
		}

		// Разворачиваем статус если есть
		if statusName.Valid {
			task.Status = &model.Status{
				ID:    statusID,
				Name:  statusName.String,
				Color: statusColor.String,
			}
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// ============================================================================
// Process-related методы (для v2)
// ============================================================================

func (r *taskRepository) GetAvailableStatuses(ctx context.Context, projectID uuid.UUID) ([]*model.Status, error) {
	query := `
		SELECT s.* FROM statuses s
		JOIN processes p ON s.process_id = p.id
		WHERE (p.project_id = ? OR p.is_default = TRUE)
		ORDER BY s.order ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*model.Status
	for rows.Next() {
		s := &model.Status{}
		err := rows.Scan(&s.ID, &s.ProcessID, &s.Name, &s.Color, &s.Order, &s.IsFinal, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}

	return statuses, nil
}

func (r *taskRepository) GetTransitionsForStatus(ctx context.Context, projectID, statusID uuid.UUID) ([]*model.Transition, error) {
	query := `
		SELECT t.* FROM transitions t
		JOIN processes p ON t.process_id = p.id
		WHERE (p.project_id = ? OR p.is_default = TRUE)
		AND (t.from_status_id = ? OR t.from_status_id = 0)
		ORDER BY t.order ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, statusID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transitions []*model.Transition
	for rows.Next() {
		tr := &model.Transition{}
		err := rows.Scan(&tr.ID, &tr.ProcessID, &tr.FromStatusID, &tr.ToStatusID, &tr.Name, &tr.Order, &tr.CreatedAt)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, tr)
	}

	return transitions, nil
}

// parseUUID — безопасно парсит UUID из строки
func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid: %w", err)
	}
	return id, nil
}

// uuidOrNull — возвращает uuid.Nil если указатель nil
func uuidOrNull(u *uuid.UUID) uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return *u
}
