package repository

import (
	"context"
	"database/sql"
	"tracker/internal/model"
)

type ProjectRepository interface {
	// Project CRUD
	Create(ctx context.Context, project *model.Project) (int64, error)
	GetByID(ctx context.Context, id int) (*model.Project, error)
	GetAll(ctx context.Context, userID int) ([]model.Project, error)
	Update(ctx context.Context, project *model.Project) error
	Delete(ctx context.Context, id int) error

	// Project Member management
	GetMember(ctx context.Context, projectID, userID int) (*model.ProjectMember, error)
	AddMember(ctx context.Context, member *model.ProjectMember) error
	RemoveMember(ctx context.Context, projectID, userID int) error
	UpdateMemberRole(ctx context.Context, projectID, userID int, role model.ProjectRole) error
	GetMembers(ctx context.Context, projectID int) ([]model.ProjectMember, error)

	// Access check helpers
	IsMember(ctx context.Context, projectID, userID int) (bool, error)
	IsAdmin(ctx context.Context, projectID, userID int) (bool, error)
}

type projectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *model.Project) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (name, description, owner_id, is_active, created_at, updated_at) 
         VALUES (?, ?, ?, 1, ?, ?)`,
		project.Name, project.Description, project.OwnerID, project.CreatedAt, project.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *projectRepository) GetByID(ctx context.Context, id int) (*model.Project, error) {
	var p model.Project
	var ownerID sql.NullInt64
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, owner_id, is_active, created_at, updated_at, deleted_at 
         FROM projects WHERE id = ? AND is_active = 1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &ownerID, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt, &deletedAt)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	p.OwnerID = ownerID
	p.DeletedAt = deletedAt
	return &p, nil
}

func (r *projectRepository) GetAll(ctx context.Context, userID int) ([]model.Project, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT p.id, p.name, p.description, p.owner_id, p.is_active, p.created_at, p.updated_at
         FROM projects p
         JOIN project_members pm ON p.id = pm.project_id
         WHERE pm.user_id = ? AND p.is_active = 1
         ORDER BY p.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		var ownerID sql.NullInt64
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &ownerID, &p.IsActive,
			&p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		p.OwnerID = ownerID
		projects = append(projects, p)
	}

	return projects, rows.Err()
}

func (r *projectRepository) Update(ctx context.Context, project *model.Project) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE projects SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		project.Name, project.Description, project.UpdatedAt, project.ID,
	)
	return err
}

func (r *projectRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE projects SET is_active = 0, deleted_at = ? WHERE id = ?`,
		sql.NullTime{Time: sql.NullTime{}.Time, Valid: true}, id,
	)
	return err
}

func (r *projectRepository) GetMember(ctx context.Context, projectID, userID int) (*model.ProjectMember, error) {
	var m model.ProjectMember
	var addedBy sql.NullInt64

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

func (r *projectRepository) AddMember(ctx context.Context, member *model.ProjectMember) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO project_members (project_id, user_id, role, added_by, joined_at) 
         VALUES (?, ?, ?, ?, ?)
         ON CONFLICT(project_id, user_id) DO UPDATE SET role = excluded.role, added_by = excluded.added_by`,
		member.ProjectID, member.UserID, member.Role, member.AddedBy, member.JoinedAt,
	)
	return err
}

func (r *projectRepository) RemoveMember(ctx context.Context, projectID, userID int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM project_members WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	)
	return err
}

func (r *projectRepository) UpdateMemberRole(ctx context.Context, projectID, userID int, role model.ProjectRole) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE project_members SET role = ? WHERE project_id = ? AND user_id = ?`,
		role, projectID, userID,
	)
	return err
}

func (r *projectRepository) GetMembers(ctx context.Context, projectID int) ([]model.ProjectMember, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at, pm.added_by, u.email
         FROM project_members pm
         JOIN users u ON pm.user_id = u.id
         WHERE pm.project_id = ?
         ORDER BY pm.joined_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.ProjectMember
	for rows.Next() {
		var m model.ProjectMember
		var addedBy sql.NullInt64
		err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.JoinedAt, &addedBy, &m.UserEmail)
		if err != nil {
			return nil, err
		}
		m.AddedBy = addedBy
		members = append(members, m)
	}

	return members, rows.Err()
}

func (r *projectRepository) IsMember(ctx context.Context, projectID, userID int) (bool, error) {
	member, err := r.GetMember(ctx, projectID, userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return member != nil, nil
}

func (r *projectRepository) IsAdmin(ctx context.Context, projectID, userID int) (bool, error) {
	member, err := r.GetMember(ctx, projectID, userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return member != nil && member.Role == model.RoleProjectAdmin, nil
}
