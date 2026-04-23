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
	GetByID(ctx context.Context, id int) (*model.User, error)
	UpdateLastLogin(ctx context.Context, id int) error
	Deactivate(ctx context.Context, userID, delettedBy int) error
	Reactivate(ctx context.Context, userID, reactivatedBy int) error
	IncrementTokenVersion(ctx context.Context, userID int) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (email, password_hash, created_at) 
		 VALUES (?, ?, ?)`,
		user.Email, user.PasswordHash, time.Now(),
	)
	return err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	var lastLogin, deletedAt, reactivatedAt sql.NullTime
	var deletedBy, reactivatedBy sql.NullInt64

	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, is_active, deleted_at, deleted_by, 
                created_at, last_login, token_version, require_password_reset, reactivated_at, reactivated_by
         FROM users WHERE email = ? AND is_active = 1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsActive, &deletedAt,
		&deletedBy, &u.CreatedAt, &lastLogin, &u.TokenVersion,
		&u.RequirePasswordReset, &reactivatedAt, &reactivatedBy)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	u.DeletedAt = deletedAt
	u.DeletedBy = deletedBy
	u.ReactivatedAt = reactivatedAt
	u.ReactivatedBy = reactivatedBy

	if lastLogin.Valid {
		u.LastLogin = lastLogin
	}

	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	var u model.User
	var lastLogin, deletedAt, reactivatedAt sql.NullTime
	var deletedBy, reactivatedBy sql.NullInt64

	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, is_active, deleted_at, deleted_by, 
                created_at, last_login, token_version, require_password_reset, reactivated_at, reactivated_by
         FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsActive, &deletedAt,
		&deletedBy, &u.CreatedAt, &lastLogin, &u.TokenVersion,
		&u.RequirePasswordReset, &reactivatedAt, &reactivatedBy)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	u.DeletedAt = deletedAt
	u.DeletedBy = deletedBy
	u.ReactivatedAt = reactivatedAt
	u.ReactivatedBy = reactivatedBy

	if lastLogin.Valid {
		u.LastLogin = lastLogin
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

// Deactivate is a user "soft-delete"
func (r *userRepository) Deactivate(ctx context.Context, userID, deletedBy int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active = 0, deleted_at = ?, deleted_by = ? WHERE id = ?`,
		time.Now(), deletedBy, userID,
	)
	return err
}

func (r *userRepository) Reactivate(ctx context.Context, userID, reactivatedBy int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET 
            is_active = 1, 
            deleted_at = NULL, 
            deleted_by = NULL,
            require_password_reset = 1,
            reactivated_at = ?,
            reactivated_by = ?
         WHERE id = ?`,
		time.Now(), reactivatedBy, userID,
	)
	return err
}

func (r *userRepository) IncrementTokenVersion(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET token_version = token_version + 1 WHERE id = ?`,
		userID,
	)
	return err
}
