package mocks

import (
	"context"
	"database/sql"
	"tracker/internal/model"
)

type MockTaskRepository struct {
	GetByIDFunc           func(ctx context.Context, id int) (*model.Task, error)
	CreateFunc            func(ctx context.Context, task *model.Task) (sql.Result, error)
	UpdateFunc            func(ctx context.Context, task *model.Task) error
	DeleteFunc            func(ctx context.Context, id int) error
	GetWithPaginationFunc func(ctx context.Context, limit, offset int) ([]model.Task, error)
	CountFunc             func(ctx context.Context) (int, error)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id int) (*model.Task, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}

	return nil, sql.ErrNoRows
}

func (m *MockTaskRepository) Create(ctx context.Context, task *model.Task) (sql.Result, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, task)
	}

	return nil, nil
}

func (m *MockTaskRepository) Update(ctx context.Context, task *model.Task) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, task)
	}
	return nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockTaskRepository) GetWithPagination(ctx context.Context, limit, offset int) ([]model.Task, error) {
	if m.GetWithPaginationFunc != nil {
		return m.GetWithPaginationFunc(ctx, limit, offset)
	}
	return []model.Task{}, nil
}

func (m *MockTaskRepository) Count(ctx context.Context) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return 0, nil
}
