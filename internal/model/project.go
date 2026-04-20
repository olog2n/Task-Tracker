package model

import (
	"database/sql"
	"time"
)

type ProjectRole string

const (
	RoleProjectAdmin  ProjectRole = "admin"  // Whole access
	RoleProjectMember ProjectRole = "member" // Only Read and Write
	RoleProjectViewer ProjectRole = "viewer" // Only Read
)

type Project struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	OwnerID     sql.NullInt64 `json:"owner_id"`
	IsActive    bool          `json:"is_active"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DeletedAt   sql.NullTime  `json:"deleted_at,omitempty"`
}

type ProjectMember struct {
	ID        int           `json:"-"`
	ProjectID int           `json:"project_id"`
	UserID    int           `json:"user_id"`
	Role      ProjectRole   `json:"role"`
	JoinedAt  time.Time     `json:"joined_at"`
	AddedBy   sql.NullInt64 `json:"added_by"`
	UserEmail string        `json:"user_email,omitempty"`
}

func (m *ProjectMember) CanView() bool {
	return true
}

func (m *ProjectMember) CanEdit() bool {
	return m.Role == RoleProjectAdmin || m.Role == RoleProjectMember
}

func (m *ProjectMember) CanManageMembers() bool {
	return m.Role == RoleProjectAdmin
}

func (m *ProjectMember) CanDelete() bool {
	return m.Role == RoleProjectAdmin
}
