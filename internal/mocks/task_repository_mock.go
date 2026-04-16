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
	UpdateFunc  func(ctx context.Context, task *model.Task) error
	DeleteFunc  func(ctx context.Context, id int) error
}

func (m *MockTaskRepository) GetAll(ctx context.Context) ([]model.Task, error) {
	return m.GetAllFunc(ctx)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id int) (*model.Task, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockTaskRepository) Create(ctx context.Context, task *model.Task) (sql.Result, error) {
	return m.CreateFunc(ctx, task)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *model.Task) error {
	return m.UpdateFunc(ctx, task)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id int) error {
	return m.DeleteFunc(ctx, id)
}
