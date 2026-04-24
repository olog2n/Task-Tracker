package model

import (
	"time"

	"github.com/google/uuid"
)

type ActionType string

// ============================================================================
// ActionType — тип действия в аудите
// ============================================================================
const (
	ActionSelect       ActionType = "select"
	ActionCreate       ActionType = "create"
	ActionUpdate       ActionType = "update"
	ActionDelete       ActionType = "delete"
	ActionStatusChange ActionType = "status_change"
	ActionAssign       ActionType = "assign"
	ActionLogin        ActionType = "login"
	ActionLogout       ActionType = "logout"
	ActionExport       ActionType = "export"
	ActionSearch       ActionType = "search"
)

// ============================================================================
// AuditLog — запись в журнале аудита (соответствует БД)
// ============================================================================
type AuditLog struct {
	ID             uuid.UUID          `json:"id" db:"id"`
	ActorID        *uuid.UUID         `json:"actor_id,omitempty" db:"actor_id"`
	UserEmail      string             `json:"user_email" db:"user_email"`
	UserName       string             `json:"user_name" db:"user_name"`
	Action         ActionType         `json:"action" db:"action"`
	TargetType     string             `json:"target_type" db:"target_type"`
	TargetID       *uuid.UUID         `json:"target_id,omitempty" db:"target_id"`
	OldValue       string             `json:"old_value,omitempty" db:"old_value"`
	NewValue       string             `json:"new_value,omitempty" db:"new_value"`
	Metadata       string             `json:"metadata,omitempty" db:"metadata"`
	Classification DataClassification `json:"classification" db:"classification"`
	IPAddress      string             `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      string             `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt      time.Time          `json:"created_at" db:"created_at"`
}

// ============================================================================
// AuditInput — для создания записи (из хендлера)
// ============================================================================
type AuditInput struct {
	ActorID        *uuid.UUID
	UserEmail      string
	UserName       string
	Action         ActionType
	TargetType     string
	TargetID       *uuid.UUID
	OldValue       interface{}
	NewValue       interface{}
	Metadata       interface{}
	Classification DataClassification `json:"classification"`
	IPAddress      string
	UserAgent      string
}

// AuditFilters — фильтры для запроса аудита
type AuditFilters struct {
	TargetType     string             `json:"target_type,omitempty"`
	TargetID       *uuid.UUID         `json:"target_id,omitempty"`
	ActorID        *uuid.UUID         `json:"actor_id,omitempty"`
	Action         ActionType         `json:"action,omitempty"`
	Classification DataClassification `json:"classification,omitempty"`
	DateFrom       time.Time          `json:"date_from,omitempty"`
	DateTo         time.Time          `json:"date_to,omitempty"`
	Limit          int                `json:"limit"`
	Offset         int                `json:"offset"`
}
