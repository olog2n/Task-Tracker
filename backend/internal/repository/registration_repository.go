package repository

import (
	"context"
	"database/sql"
	"tracker/internal/model"
)

type RegistrationRepository interface {
	GetAll(ctx context.Context) ([]model.User, error)
	GetById(ctx context.Context, id int) (*model.Task, error)
	Create(ctx context.Context, user *model.User) (sql.Result, error)
}

type registrationRepository struct {
	db *sql.DB
}

func NewRegistrationRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *registrationRepository) GetAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, last_login, created_at, email from users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.LastLogin, &u.CreatedAt, &u.Email)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

// func (r *registrationRepository) GetById(ctx context.Context) ([]model.User, error) {

// }

// func (r *registrationRepository) Create(ctx context.Context, user *model.User) (sql.Result, error) {

// }
