package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TaskID int

type Task struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`

	StatusID uuid.UUID `json:"status_id" db:"status_id"`
	Status   *Status   `json:"status,omitempty" db:"-"`

	Priority   string     `json:"priority" db:"priority"`
	ProjectID  uuid.UUID  `json:"project_id" db:"project_id"`
	AssigneeID *uuid.UUID `json:"assignee_id" db:"assignee_id"`
	CreatedBy  uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedBy  *uuid.UUID `json:"updated_by" db:"updated_by"`
	DeletedBy  *uuid.UUID `json:"deleted_by,omitempty" db:"deleted_by"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	Classification DataClassification `json:"classification" db:"cassification"` // internal/confidential/restricted
	IsSensitive    bool               `json:"is_sensitive" db:"is_sensitive"`    //Для быстрого поиска, но в перспективе стоит удалить
}

// ============================================================================
// TaskInput — данные для создания/обновления задачи
// ============================================================================
type TaskInput struct {
	Title       string     `json:"title" validate:"required,min=1,max=500"`
	Description string     `json:"description" validate:"max=10000"`
	StatusID    *uuid.UUID `json:"status_id,omitempty"`
	Priority    string     `json:"priority" validate:"oneof=low medium high"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
}

// ============================================================================
// TaskFilter — фильтры для поиска задач
// ============================================================================
type TaskFilter struct {
	ProjectID      *uuid.UUID         `json:"project_id,omitempty"`
	AssigneeID     *uuid.UUID         `json:"assignee_id,omitempty"`
	StatusID       *uuid.UUID         `json:"status_id,omitempty"`
	Priority       string             `json:"priority,omitempty"`
	Search         string             `json:"search,omitempty"`
	Classification DataClassification `json:"classification,omitempty"`
	Sensitive      *bool              `json:"sensitive,omitempty"`
	SortBy         string             `json:"sort_by,omitempty"`
	SortOrder      string             `json:"sort_order,omitempty"`
	Limit          int                `json:"limit"`
	Offset         int                `json:"offset"`
}

// ============================================================================
// TaskList — ответ со списком задач (с пагинацией)
// ============================================================================
type TaskList struct {
	Tasks   []*Task `json:"tasks"`
	Total   int     `json:"total"`
	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	HasMore bool    `json:"has_more"`
}

// ============================================================================
// PaginatedTask — ответ со списком задач (с пагинацией)
// ============================================================================
type PaginatedTasks struct {
	Tasks   []*Task `json:"tasks"`
	Total   int     `json:"total"`
	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	HasMore bool    `json:"has_more"`
}

// ============================================================================
// Методы Task
// ============================================================================

func (t *Task) GetDataFields() map[string]DataField {
	return map[string]DataField{
		"title":       {Name: "title", DataType: DataTypeBusiness, Classification: t.Classification},
		"description": {Name: "description", DataType: DataTypeBusiness, Classification: t.Classification},
		"assignee_id": {Name: "assignee_id", DataType: DataTypePersonal, Classification: ClassificationConfidential},
		"priority":    {Name: "priority", DataType: DataTypeBusiness, Classification: ClassificationInternal},
	}
}

// GetAuditLevel — возвращает уровень классификации для аудита
func (t *Task) GetAuditLevel() DataClassification {
	// Если задача помечена как чувствительная — сразу confidential
	if t.IsSensitive {
		return ClassificationConfidential
	}

	// Если есть назначенный исполнитель — минимум confidential (ПДн)
	if t.AssigneeID != nil {
		return ClassificationConfidential
	}

	// Иначе используем классификацию задачи
	return t.Classification
}

// GetStatusName — удобное получение имени статуса
func (t *Task) GetStatusName() string {
	if t.Status != nil {
		return t.Status.Name
	}
	return ""
}

// GetStatusColor — удобное получение цвета статуса (для UI)
func (t *Task) GetStatusColor() string {
	if t.Status != nil {
		return t.Status.Color
	}
	return "#6b7280" // default gray
}

func (ti TaskID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(ti))
}

func (ti *TaskID) UnmarshalJSON(data []byte) error {
	var res int
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	*ti = TaskID(res)

	return nil
}
