package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"tracker/internal/model"
	"tracker/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupUserRepo(t *testing.T) (*repository.UserRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	repo := repository.NewUserRepository(db)

	return &repo, mock, func() {
		db.Close()
	}
}

func TestUserRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.Email, user.PasswordHash, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := (*repo).Create(ctx, user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo, mock, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	email := "test@example.com"

	rows := sqlmock.NewRows([]string{
		"id", "email", "password_hash", "is_active", "deleted_at", "deleted_by",
		"created_at", "last_login", "token_version", "require_password_reset",
		"reactivated_at", "reactivated_by",
	}).AddRow(
		1, email, "hashed", true, nil, nil,
		time.Now(), nil, 1, false,
		nil, nil,
	)

	mock.ExpectQuery("SELECT .* FROM users WHERE email = \\? AND is_active = 1").
		WithArgs(email).
		WillReturnRows(rows)

	user, err := (*repo).GetByEmail(ctx, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	email := "notfound@example.com"

	mock.ExpectQuery("SELECT .* FROM users WHERE email = \\? AND is_active = 1").
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	user, err := (*repo).GetByEmail(ctx, email)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Deactivate(t *testing.T) {
	repo, mock, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	userID := 1
	deletedBy := 2

	mock.ExpectExec("UPDATE users SET is_active = 0, deleted_at = \\?, deleted_by = \\? WHERE id = \\?").
		WithArgs(sqlmock.AnyArg(), deletedBy, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := (*repo).Deactivate(ctx, userID, deletedBy)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Reactivate(t *testing.T) {
	repo, mock, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	userID := 1
	reactivatedBy := 2

	mock.ExpectExec("UPDATE users SET is_active = 1, deleted_at = NULL").
		WithArgs(sqlmock.AnyArg(), reactivatedBy, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := (*repo).Reactivate(ctx, userID, reactivatedBy)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_IncrementTokenVersion(t *testing.T) {
	repo, mock, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	userID := 1

	mock.ExpectExec("UPDATE users SET token_version = token_version \\+ 1 WHERE id = \\?").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := (*repo).IncrementTokenVersion(ctx, userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
