package repository

import (
	"context"
	"database/sql"
	"time"
	"tracker/internal/model"
)

type TaskRepository interface {
	GetAll(ctx context.Context) ([]model.Task, error)
	GetByID(ctx context.Context, id int) (*model.Task, error)
	Create(ctx context.Context, task *model.Task) (sql.Result, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) GetAll(ctx context.Context) ([]model.Task, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, author, description, executor, status, created_at, updated_at FROM tasks`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		var statusStr string
		var createdAt, updatedAt time.Time
		err := rows.Scan(&t.ID, &t.Title, &t.Author, &t.Description, &t.Executor, &statusStr, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		t.Status, _ = model.FromString(statusStr)
		t.CreatedAt = createdAt
		t.UpdatedAt = updatedAt
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id int) (*model.Task, error) {
	var t model.Task
	var statusStr string
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, author, description, executor, status, created_at, updated_at FROM tasks WHERE id = ?`,
		id,
	).Scan(&t.ID, &t.Title, &t.Author, &t.Description, &t.Executor, &statusStr, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	t.Status, _ = model.FromString(statusStr)
	t.CreatedAt = createdAt
	t.UpdatedAt = updatedAt
	return &t, nil
}

func (r *taskRepository) Create(ctx context.Context, task *model.Task) (sql.Result, error) {
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	return r.db.ExecContext(ctx,
		`INSERT INTO tasks (title, author, description, executor, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.Title, task.Author, task.Description, task.Executor, task.Status.ToString(), now, now,
	)
}
