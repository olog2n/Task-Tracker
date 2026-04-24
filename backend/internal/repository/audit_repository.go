package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"tracker/internal/model"

	"github.com/google/uuid"
)

type AuditRepository interface {
	Log(ctx context.Context, input *model.AuditInput) error
	GetByTarget(ctx context.Context, targetType string, targetID uuid.UUID, limit int) ([]*model.AuditLog, error)
	GetByActor(ctx context.Context, actorID uuid.UUID, limit int) ([]*model.AuditLog, error)
	GetWithFilters(ctx context.Context, filters *model.AuditFilters) ([]*model.AuditLog, error)
}

type auditRepo struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) AuditRepository {
	return &auditRepo{db: db}
}

// ============================================================================
// Log — запись новой записи аудита
// ============================================================================
func (r *auditRepo) Log(ctx context.Context, input *model.AuditInput) error {
	// Сериализуем JSON поля
	var oldValue, newValue, metadata string
	if input.OldValue != nil {
		b, err := json.Marshal(input.OldValue)
		if err != nil {
			oldValue = "{}"
		} else {
			oldValue = string(b)
		}
	}
	if input.NewValue != nil {
		b, err := json.Marshal(input.NewValue)
		if err != nil {
			newValue = "{}"
		} else {
			newValue = string(b)
		}
	}
	if input.Metadata != nil {
		b, err := json.Marshal(input.Metadata)
		if err != nil {
			metadata = "{}"
		} else {
			metadata = string(b)
		}
	}

	query := `
		INSERT INTO audit_log 
		(actor_id, user_email, user_name, action, target_type, target_id, 
		 old_value, new_value, metadata, classification, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var actorIDStr, targetIDStr *string
	if input.ActorID != nil {
		s := input.ActorID.String()
		actorIDStr = &s
	}
	if input.TargetID != nil {
		s := input.TargetID.String()
		targetIDStr = &s
	}

	_, err := r.db.ExecContext(ctx, query,
		actorIDStr,
		input.UserEmail,
		input.UserName,
		input.Action,
		input.TargetType,
		targetIDStr,
		oldValue,
		newValue,
		metadata,
		string(input.Classification),
		input.IPAddress,
		input.UserAgent)

	return err
}

// ============================================================================
// GetByTarget — получить аудит по сущности (например, история задачи)
// ============================================================================
func (r *auditRepo) GetByTarget(ctx context.Context, targetType string, targetID uuid.UUID, limit int) ([]*model.AuditLog, error) {
	query := `
		SELECT * FROM audit_log 
		WHERE target_type = ? AND target_id = ? 
		ORDER BY created_at DESC 
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, targetType, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ============================================================================
// GetByActor — получить аудит по пользователю (активность)
// ============================================================================
func (r *auditRepo) GetByActor(ctx context.Context, actorID uuid.UUID, limit int) ([]*model.AuditLog, error) {
	query := `
		SELECT * FROM audit_log 
		WHERE actor_id = ? 
		ORDER BY created_at DESC 
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, actorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ============================================================================
// GetWithFilters — получить аудит с фильтрами (админка)
// ============================================================================
func (r *auditRepo) GetWithFilters(ctx context.Context, filters *model.AuditFilters) ([]*model.AuditLog, error) {
	if filters == nil {
		filters = &model.AuditFilters{
			Limit:  100,
			Offset: 0,
		}
	}

	// Базовый запрос
	query := `SELECT * FROM audit_log WHERE 1=1`
	args := []interface{}{}

	// Динамически добавляем фильтры
	if filters.TargetType != "" {
		query += " AND target_type = ?"
		args = append(args, filters.TargetType)
	}
	if filters.TargetID != nil {
		query += " AND target_id = ?"
		args = append(args, *filters.TargetID)
	}
	if filters.ActorID != nil {
		query += " AND actor_id = ?"
		args = append(args, *filters.ActorID)
	}
	if filters.Action != "" {
		query += " AND action = ?"
		args = append(args, string(filters.Action))
	}
	if filters.Classification != "" {
		query += " AND classification = ?"
		args = append(args, string(filters.Classification))
	}
	if !filters.DateFrom.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filters.DateFrom)
	}
	if !filters.DateTo.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filters.DateTo)
	}

	// Сортировка и лимит
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ============================================================================
// Helper: scanAuditLogs — сканирует rows в []*AuditLog
// ============================================================================
func (r *auditRepo) scanAuditLogs(rows *sql.Rows) ([]*model.AuditLog, error) {
	var logs []*model.AuditLog
	for rows.Next() {
		log := &model.AuditLog{}
		var actorIDStr, targetIDStr, classificationStr sql.NullString

		err := rows.Scan(
			&log.ID, // uuid.UUID
			&actorIDStr,
			&log.UserEmail,
			&log.UserName,
			&log.Action,
			&log.TargetType,
			&targetIDStr,
			&log.OldValue,
			&log.NewValue,
			&log.Metadata,
			&classificationStr,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if actorIDStr.Valid {
			id, err := uuid.Parse(actorIDStr.String)
			if err == nil {
				log.ActorID = &id
			}
		}
		if targetIDStr.Valid {
			id, err := uuid.Parse(targetIDStr.String)
			if err == nil {
				log.TargetID = &id
			}
		}
		if classificationStr.Valid {
			log.Classification = model.DataClassification(classificationStr.String)
		}

		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

// ============================================================================
// CountWithFilters — получить количество записей с фильтрами (для пагинации)
// ============================================================================
func (r *auditRepo) CountWithFilters(ctx context.Context, filters *model.AuditFilters) (int, error) {
	if filters == nil {
		filters = &model.AuditFilters{}
	}

	query := `SELECT COUNT(*) FROM audit_log WHERE 1=1`
	args := []interface{}{}

	if filters.TargetType != "" {
		query += " AND target_type = ?"
		args = append(args, filters.TargetType)
	}
	if filters.TargetID != nil {
		query += " AND target_id = ?"
		args = append(args, *filters.TargetID)
	}
	if filters.ActorID != nil {
		query += " AND actor_id = ?"
		args = append(args, *filters.ActorID)
	}
	if filters.Action != "" {
		query += " AND action = ?"
		args = append(args, string(filters.Action))
	}
	if filters.Classification != "" {
		query += " AND classification = ?"
		args = append(args, string(filters.Classification))
	}
	if !filters.DateFrom.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filters.DateFrom)
	}
	if !filters.DateTo.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filters.DateTo)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}
