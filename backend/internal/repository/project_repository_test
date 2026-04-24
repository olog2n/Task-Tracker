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

func setupProjectRepo(t *testing.T) (*repository.ProjectRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	repo := repository.NewProjectRepository(db)

	return &repo, mock, func() {
		db.Close()
	}
}

func TestProjectRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := &model.Project{
		Name:        "Test Project",
		Description: "Test Description",
		OwnerID:     sql.NullInt64{Int64: 1, Valid: true},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mock.ExpectExec("INSERT INTO projects").
		WithArgs(project.Name, project.Description, project.OwnerID.Int64, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	projectID, err := (*repo).Create(ctx, project)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), projectID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()
	projectID := 1

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "owner_id", "is_active", "created_at", "updated_at", "deleted_at",
	}).AddRow(
		projectID, "Test Project", "Test Description", 1, true, time.Now(), time.Now(), nil,
	)

	mock.ExpectQuery("SELECT .* FROM projects WHERE id = \\? AND is_active = 1").
		WithArgs(projectID).
		WillReturnRows(rows)

	project, err := (*repo).GetByID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "Test Project", project.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_GetMember(t *testing.T) {
	repo, mock, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()
	projectID := 1
	userID := 2

	rows := sqlmock.NewRows([]string{
		"id", "project_id", "user_id", "role", "joined_at", "added_by",
	}).AddRow(
		1, projectID, userID, "admin", time.Now(), sql.NullInt64{Int64: 1, Valid: true},
	)

	mock.ExpectQuery("SELECT .* FROM project_members WHERE project_id = \\? AND user_id = \\?").
		WithArgs(projectID, userID).
		WillReturnRows(rows)

	member, err := (*repo).GetMember(ctx, projectID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, model.RoleProjectAdmin, member.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_AddMember(t *testing.T) {
	repo, mock, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()
	member := &model.ProjectMember{
		ProjectID: 1,
		UserID:    2,
		Role:      model.RoleProjectMember,
		JoinedAt:  time.Now(),
		AddedBy:   sql.NullInt64{Int64: 1, Valid: true},
	}

	mock.ExpectExec("INSERT INTO project_members").
		WithArgs(member.ProjectID, member.UserID, member.Role, member.AddedBy.Int64, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := (*repo).AddMember(ctx, member)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_IsMember(t *testing.T) {
	repo, mock, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()
	projectID := 1
	userID := 2

	// Тест: пользователь является участником
	rows := sqlmock.NewRows([]string{"id", "project_id", "user_id", "role", "joined_at", "added_by"}).
		AddRow(1, projectID, userID, "member", time.Now(), nil)

	mock.ExpectQuery("SELECT .* FROM project_members").
		WithArgs(projectID, userID).
		WillReturnRows(rows)

	isMember, err := (*repo).IsMember(ctx, projectID, userID)

	assert.NoError(t, err)
	assert.True(t, isMember)

	// Тест: пользователь не является участником
	mock.ExpectQuery("SELECT .* FROM project_members").
		WithArgs(projectID, userID).
		WillReturnError(sql.ErrNoRows)

	isMember, err = (*repo).IsMember(ctx, projectID, userID)

	assert.NoError(t, err)
	assert.False(t, isMember)
	assert.NoError(t, mock.ExpectationsWereMet())
}
