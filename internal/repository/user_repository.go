package repository

import (
	"context"
	"database/sql"
	"time"
	"tracker/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateLastLogin(ctx context.Context, id int) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (email, password_hash, created_at) VALUES (?, ?, ?)`,
		user.Email, user.PasswordHash, time.Now(),
	)
	return err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	var lastLogin sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, created_at, last_login FROM users WHERE email = ?`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &lastLogin)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		u.LastLogin = lastLogin.Time
	}

	return &u, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}
