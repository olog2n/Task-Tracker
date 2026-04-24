package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"tracker/internal/model"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	// Projects
	CreateProject(ctx context.Context, project *model.Project) (uuid.UUID, error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	GetProjectsByOwner(ctx context.Context, ownerID uuid.UUID) ([]*model.Project, error)
	GetAllProjects(ctx context.Context, limit, offset int) ([]*model.Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, input *model.ProjectInput, updatedBy uuid.UUID) (*model.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error

	// Project members
	AddProjectMember(ctx context.Context, member *model.ProjectMember) error
	GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*model.ProjectMember, error)
	GetMemberByID(ctx context.Context, projectID, userID uuid.UUID) (*model.ProjectMember, error)
	UpdateMemberRole(ctx context.Context, projectID, userID uuid.UUID, role model.ProjectRole) error
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error

	// // Access check helpers
	IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
	IsAdmin(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
}

type projectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) CreateProject(ctx context.Context, project *model.Project) (uuid.UUID, error) {
	// Генерируем UUID если не задан
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}

	query := `
		INSERT INTO projects 
		(id, name, description, owner_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	_, err := r.db.ExecContext(ctx, query,
		project.ID.String(),
		project.Name,
		project.Description,
		project.OwnerID.String(),
		now,
		now,
	)

	if err != nil {
		return uuid.Nil, err
	}

	project.CreatedAt = now
	project.UpdatedAt = now

	return project.ID, nil
}

func (r *projectRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM projects
		WHERE id = ? AND deleted_at IS NULL
	`

	project := &model.Project{}
	var idStr, ownerIDStr string

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&project.Name,
		&project.Description,
		&ownerIDStr,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Конвертируем string → UUID
	project.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid project id: %w", err)
	}

	project.OwnerID, err = uuid.Parse(ownerIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid owner id: %w", err)
	}

	return project, nil
}

func (r *projectRepository) GetProjectsByOwner(ctx context.Context, ownerID uuid.UUID) ([]*model.Project, error) {
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM projects
		WHERE owner_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	return r.scanProjects(query, ownerID.String())
}

func (r *projectRepository) GetAllProjects(ctx context.Context, limit, offset int) ([]*model.Project, error) {
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM projects
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	return r.scanProjects(query, limit, offset)
}

func (r *projectRepository) UpdateProject(ctx context.Context, id uuid.UUID, input *model.ProjectInput, updatedBy uuid.UUID) (*model.Project, error) {
	// Получаем текущий проект
	oldProject, err := r.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Строим динамический UPDATE
	updates := []string{}
	args := []interface{}{}

	if input.Name != "" && input.Name != oldProject.Name {
		updates = append(updates, "name = ?")
		args = append(args, input.Name)
	}

	if input.Description != oldProject.Description {
		updates = append(updates, "description = ?")
		args = append(args, input.Description)
	}

	if input.OwnerID != nil && *input.OwnerID != oldProject.OwnerID {
		updates = append(updates, "owner_id = ?")
		args = append(args, input.OwnerID.String())
	}

	// Всегда обновляем updated_at
	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())

	if len(updates) == 0 {
		return oldProject, nil
	}

	query := fmt.Sprintf(`UPDATE projects SET %s WHERE id = ?`, joinStrings(updates, ", "))
	args = append(args, id.String())

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return r.GetProjectByID(ctx, id)
}

func (r *projectRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE projects SET deleted_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id.String())
	return err
}

func (r *projectRepository) AddProjectMember(ctx context.Context, member *model.ProjectMember) error {
	// Генерируем UUID если не задан
	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}

	query := `
		INSERT INTO project_members 
		(id, project_id, user_id, role, joined_at, added_by, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		member.ID.String(),
		member.ProjectID.String(),
		member.UserID.String(),
		member.Role,
		now,
		member.AddedBy.String(),
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to add project member: %w", err)
	}

	member.JoinedAt = now
	member.UpdatedAt = now

	return nil
}

func (r *projectRepository) GetMember(ctx context.Context, projectID, userID uuid.UUID) (*model.ProjectMember, error) {
	var m model.ProjectMember
	var addedBy uuid.UUID

	err := r.db.QueryRowContext(ctx,
		`SELECT id, project_id, user_id, role, joined_at, added_by 
         FROM project_members WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	).Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.JoinedAt, &addedBy)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	m.AddedBy = addedBy
	return &m, nil
}

func (r *projectRepository) GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*model.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, joined_at, added_by, updated_at
		FROM project_members
		WHERE project_id = ?
		ORDER BY joined_at DESC
	`

	return r.scanProjectMembers(query, projectID.String())
}

func (r *projectRepository) GetMemberByID(ctx context.Context, projectID, userID uuid.UUID) (*model.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, joined_at, added_by, updated_at
		FROM project_members
		WHERE project_id = ? AND user_id = ?
	`

	member := &model.ProjectMember{}
	var idStr, projectIDStr, userIDStr, addedByStr string

	err := r.db.QueryRowContext(ctx, query, projectID.String(), userID.String()).Scan(
		&idStr,
		&projectIDStr,
		&userIDStr,
		&member.Role,
		&member.JoinedAt,
		&addedByStr,
		&member.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("member not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	// Конвертируем string → UUID
	member.ID, _ = uuid.Parse(idStr)
	member.ProjectID, _ = uuid.Parse(projectIDStr)
	member.UserID, _ = uuid.Parse(userIDStr)
	member.AddedBy, _ = uuid.Parse(addedByStr)

	return member, nil
}

func (r *projectRepository) UpdateMemberRole(ctx context.Context, projectID, userID uuid.UUID, role model.ProjectRole) error {
	query := `UPDATE project_members SET role = ?, updated_at = ? WHERE project_id = ? AND user_id = ?`

	_, err := r.db.ExecContext(ctx, query, role, time.Now(), projectID.String(), userID.String())
	return err
}

func (r *projectRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	query := `DELETE FROM project_members WHERE project_id = ? AND user_id = ?`

	_, err := r.db.ExecContext(ctx, query, projectID.String(), userID.String())
	return err
}

func (r *projectRepository) IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	member, err := r.GetMember(ctx, projectID, userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return member != nil, nil
}

func (r *projectRepository) IsAdmin(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	member, err := r.GetMember(ctx, projectID, userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return member != nil && member.Role == model.RoleProjectAdmin, nil
}

func (r *projectRepository) scanProjects(query string, args ...interface{}) ([]*model.Project, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*model.Project
	for rows.Next() {
		p := &model.Project{}
		var idStr, ownerIDStr string

		err := rows.Scan(&idStr, &p.Name, &p.Description, &ownerIDStr, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}

		p.ID, _ = uuid.Parse(idStr)
		p.OwnerID, _ = uuid.Parse(ownerIDStr)

		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *projectRepository) scanProjectMembers(query string, args ...interface{}) ([]*model.ProjectMember, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.ProjectMember
	for rows.Next() {
		m := &model.ProjectMember{}
		var idStr, projectIDStr, userIDStr, addedByStr string

		err := rows.Scan(&idStr, &projectIDStr, &userIDStr, &m.Role, &m.JoinedAt, &addedByStr, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}

		m.ID, _ = uuid.Parse(idStr)
		m.ProjectID, _ = uuid.Parse(projectIDStr)
		m.UserID, _ = uuid.Parse(userIDStr)
		m.AddedBy, _ = uuid.Parse(addedByStr)

		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
