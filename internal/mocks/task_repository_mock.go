package mocks

import (
	"context"
	"database/sql"
	"tracker/internal/model"
)

type MockTaskRepository struct {
	GetAllFunc  func(ctx context.Context) ([]model.Task, error)
	GetByIDFunc func(ctx context.Context, id int) (*model.Task, error)
	CreateFunc  func(ctx context.Context, task *model.Task) (sql.Result, error)
}

func (m *MockTaskRepository) GetAll(ctx context.Context) ([]model.Task, error) {
	return m.GetAllFunc(ctx)
}

func (m *MockTaskRepository) GetById(ctx context.Context, id int) (*model.Task, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockTaskRepository) Create(ctx context.Context, task *model.Task) (sql.Result, error) {
	return m.CreateFunc(ctx, task)
}
