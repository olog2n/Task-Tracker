package model

import (
	"time"

	"github.com/google/uuid"
)

type ProjectRole string

const (
	RoleProjectAdmin  ProjectRole = "admin"  // Whole access
	RoleProjectMember ProjectRole = "member" // Only Read and Write
	RoleProjectViewer ProjectRole = "viewer" // Only Read
)

// ============================================================================
// Project — проект в системе
// ============================================================================
type Project struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	OwnerID     uuid.UUID  `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ProjectInput — данные для создания/обновления проекта
type ProjectInput struct {
	Name        string     `json:"name" validate:"required,min=1,max=200"`
	Description string     `json:"description" validate:"max=5000"`
	OwnerID     *uuid.UUID `json:"owner_id,omitempty"`
}

// ============================================================================
// ProjectMember — участник проекта
// ============================================================================
type ProjectMember struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	ProjectID uuid.UUID   `json:"project_id" db:"project_id"`
	UserID    uuid.UUID   `json:"user_id" db:"user_id"`
	Role      ProjectRole `json:"role" db:"role"`
	JoinedAt  time.Time   `json:"joined_at" db:"joined_at"`
	AddedBy   uuid.UUID   `json:"added_by" db:"added_by"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
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
