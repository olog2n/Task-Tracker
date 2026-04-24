package model

import (
	"time"

	"github.com/google/uuid"
)

// var statusName = map[TaskStatus]string{
// 	StatusBacklog:    "backlog",
// 	StatusInProgress: "in_progress",
// 	StatusReview:     "review",
// 	StatusDone:       "done",
// 	StatusCancelled:  "cancelled",
// }

// Status — статус в процессе
type Status struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProcessID uuid.UUID `json:"process_id" db:"process_id"`
	Name      string    `json:"name" db:"name"`         // "В работе", "Ревью"
	Color     string    `json:"color" db:"color"`       // "#3b82f6" для UI
	Order     int       `json:"order" db:"order"`       // Порядок отображения
	IsFinal   bool      `json:"is_final" db:"is_final"` // Конечный статус
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
