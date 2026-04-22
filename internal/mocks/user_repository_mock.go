package mocks

import (
	"context"
	"database/sql"
	"tracker/internal/model"
)

type MockUserRepository struct {
	CreateFunc                func(ctx context.Context, user *model.User) error
	GetByEmailFunc            func(ctx context.Context, email string) (*model.User, error)
	GetByIDFunc               func(ctx context.Context, id int) (*model.User, error)
	UpdateLastLoginFunc       func(ctx context.Context, id int) error
	DeactivateFunc            func(ctx context.Context, userID, deletedBy int) error
	ReactivateFunc            func(ctx context.Context, userID, reactivatedBy int) error
	IncrementTokenVersionFunc func(ctx context.Context, userID int) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id int) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepository) Deactivate(ctx context.Context, userID, deletedBy int) error {
	if m.DeactivateFunc != nil {
		return m.DeactivateFunc(ctx, userID, deletedBy)
	}
	return nil
}

func (m *MockUserRepository) Reactivate(ctx context.Context, userID, reactivatedBy int) error {
	if m.ReactivateFunc != nil {
		return m.ReactivateFunc(ctx, userID, reactivatedBy)
	}
	return nil
}

func (m *MockUserRepository) IncrementTokenVersion(ctx context.Context, userID int) error {
	if m.IncrementTokenVersionFunc != nil {
		return m.IncrementTokenVersionFunc(ctx, userID)
	}
	return nil
}
