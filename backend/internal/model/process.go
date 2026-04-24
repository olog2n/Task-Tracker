package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Process — шаблон процесса (workflow)
type Process struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"` // "Разработка", "Поддержка"
	Description string        `json:"description" db:"description"`
	IsDefault   bool          `json:"is_default" db:"is_default"`
	ProjectID   *uuid.UUID    `json:"project_id,omitempty" db:"project_id"` // nil = глобальный
	CreatedBy   sql.NullInt64 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// Transition — переход между статусами
type Transition struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ProcessID    uuid.UUID `json:"process_id" db:"process_id"`
	FromStatusID uuid.UUID `json:"from_status_id" db:"from_status_id"` // 0 = из любого
	ToStatusID   uuid.UUID `json:"to_status_id" db:"to_status_id"`
	Name         string    `json:"name" db:"name"` // "Начать работу"
	Order        int       `json:"order" db:"order"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
