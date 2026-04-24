package audit

import (
	"time"
	"tracker/internal/model"

	"github.com/google/uuid"
)

// ============================================================================
// AuditFilters — фильтры для запроса аудита
// ============================================================================
//
// Используется в Logger.GetAuditLogWithFilters()
// Все поля опциональны (указывай только нужные)
// ============================================================================
type AuditFilters struct {
	// Фильтр по сущности
	TargetType string     `json:"target_type,omitempty"` // "task", "project", "user"
	TargetID   *uuid.UUID `json:"target_id,omitempty"`   // ID сущности

	// Фильтр по актору (кто выполнил действие)
	ActorID *uuid.UUID `json:"actor_id,omitempty"` // ID пользователя

	// Фильтр по типу действия
	Action model.ActionType `json:"action,omitempty"` // "create", "update", "delete"

	// Фильтр по классификации данных
	Classification model.DataClassification `json:"classification,omitempty"`

	// Фильтр по времени
	DateFrom time.Time `json:"date_from,omitempty"` // >=
	DateTo   time.Time `json:"date_to,omitempty"`   // <=

	// Пагинация
	Limit  int `json:"limit"`  // Default: 100, Max: 1000
	Offset int `json:"offset"` // Default: 0
}

// Validate — валидация фильтров (опционально, можно вызвать перед запросом)
func (f *AuditFilters) Validate() error {
	if f.Limit <= 0 {
		f.Limit = 100
	}
	if f.Limit > 1000 {
		f.Limit = 1000
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return nil
}
