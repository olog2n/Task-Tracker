package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"tracker/internal/config"
	"tracker/internal/model"
	"tracker/internal/repository"

	"github.com/google/uuid"
)

// ============================================================================
// Logger — основной аудитор системы
// ============================================================================
//
// Отвечает за:
// • Проверку политик аудита (что логировать)
// • Маскирование чувствительных данных
// • Запись в audit_log через repository
// • Удобные wrapper-методы для распространённых сценариев
//
// Пример использования:
//
//	auditLogger.LogTaskViewed(ctx, r, user, task)
//	auditLogger.LogTaskUpdated(ctx, r, user, oldTask, newTask)
//	auditLogger.LogUserLogin(ctx, r, user, true, "")
//
// ============================================================================
type Logger struct {
	repo     repository.AuditRepository
	config   *config.AuditConfig
	policies map[model.DataClassification]config.ClassificationPolicy
}

// NewLogger — создаёт новый экземпляр аудит-логгера
func NewLogger(
	repo repository.AuditRepository,
	cfg *config.Config,
) *Logger {
	return &Logger{
		repo:     repo,
		config:   &cfg.Audit,
		policies: cfg.Classification.Policies,
	}
}

// ============================================================================
// Core: Базовый метод логирования
// ============================================================================

// Log — универсальный метод для записи аудита
//
// Параметры:
//   - ctx: контекст запроса
//   - r: HTTP-запрос (для IP, User-Agent)
//   - user: текущий пользователь (может быть nil для системных действий)
//   - action: тип действия (select/create/update/delete/export)
//   - targetType: тип сущности ("task", "project", "user")
//   - targetID: ID сущности (nil для массовых операций)
//   - classification: уровень чувствительности данных
//   - oldValue: старые значения (для update/delete)
//   - newValue: новые значения (для create/update)
//   - metadata: дополнительный контекст (transition_id, filters, etc.)
//
// Возвращает ошибку только если запись в БД не удалась.
// Ошибки логирования НЕ должны ломать основной поток запроса!
func (l *Logger) Log(
	ctx context.Context,
	r *http.Request,
	user *model.User,
	action model.ActionType,
	targetType string,
	targetID *uuid.UUID,
	classification model.DataClassification,
	oldValue, newValue interface{},
	metadata map[string]interface{},
) error {
	// Шаг 1: Проверка — включён ли аудит вообще
	if !l.config.Enabled {
		return nil
	}

	// Шаг 2: Проверка политики для уровня классификации
	policy := l.getPolicyForLevel(classification)
	if policy == nil {
		// Нет политики — используем дефолт (internal)
		policy = l.getPolicyForLevel(model.ClassificationInternal)
	}

	// Шаг 3: Проверяем нужно ли логировать это действие
	if !l.shouldLog(action, policy) {
		return nil
	}

	// Шаг 4: Проверяем порог для массовых select-операций
	if action == model.ActionSelect && metadata != nil {
		if count, ok := metadata["count"].(int); ok {
			if count < l.config.SelectBatchThreshold {
				return nil // Не логируем мелкое чтение
			}
		}
	}

	// Шаг 5: Маскируем данные если требуется политикой
	if policy.MaskInLogs {
		oldValue = l.maskSensitiveData(oldValue, classification)
		newValue = l.maskSensitiveData(newValue, classification)
	}

	// Шаг 6: Собираем данные для записи
	var actorID *uuid.UUID
	userEmail := "system"
	userName := "system"

	if user != nil {
		actorID = &user.ID
		userEmail = user.Email
		userName = user.Email // Или user.Name если есть
	}

	input := &model.AuditInput{
		ActorID:        actorID,
		UserEmail:      userEmail,
		UserName:       userName,
		Action:         action,
		TargetType:     targetType,
		TargetID:       targetID,
		OldValue:       oldValue,
		NewValue:       newValue,
		Metadata:       metadata,
		IPAddress:      l.getClientIP(r),
		UserAgent:      r.UserAgent(),
		Classification: classification,
	}

	// Шаг 7: Записываем в БД
	// Важно: ошибка не должна ломать запрос!
	return l.repo.Log(ctx, input)
}

// ============================================================================
// Helpers: Проверка политик
// ============================================================================

func (l *Logger) getPolicyForLevel(level model.DataClassification) *config.ClassificationPolicy {
	if policy, ok := l.policies[level]; ok {
		return &policy
	}
	return nil
}

func (l *Logger) shouldLog(action model.ActionType, policy *config.ClassificationPolicy) bool {
	if policy == nil {
		return false
	}

	switch action {
	case model.ActionSelect:
		return policy.LogSelect && l.config.LogSelect
	case model.ActionCreate:
		return policy.LogUpdate && l.config.LogUpdate
	case model.ActionUpdate:
		return policy.LogUpdate && l.config.LogUpdate
	case model.ActionDelete:
		return policy.LogDelete && l.config.LogDelete
	case model.ActionExport:
		return policy.LogExport && l.config.LogExport
	case model.ActionStatusChange:
		return policy.LogUpdate && l.config.LogUpdate
	case model.ActionAssign:
		return policy.LogUpdate && l.config.LogUpdate
	case model.ActionLogin, model.ActionLogout:
		return true // Всегда логируем аутентификацию
	default:
		return false
	}
}

// ============================================================================
// Helpers: Маскирование данных
// ============================================================================

// maskSensitiveData — маскирует чувствительные поля в зависимости от уровня
func (l *Logger) maskSensitiveData(data interface{}, level model.DataClassification) interface{} {
	if level != model.ClassificationConfidential && level != model.ClassificationRestricted {
		return data // Public/Internal не маскируем
	}

	// Для restricted — полная маскировка
	if level == model.ClassificationRestricted {
		return map[string]interface{}{
			"masked":         true,
			"classification": string(level),
			"reason":         "restricted_data",
		}
	}

	// Для confidential — частичная маскировка
	// Пытаемся замаскировать конкретные поля
	return l.maskConfidentialFields(data)
}

// maskConfidentialFields — маскирует конкретные чувствительные поля
func (l *Logger) maskConfidentialFields(data interface{}) interface{} {
	// Если это map — маскируем ключи
	if m, ok := data.(map[string]interface{}); ok {
		masked := make(map[string]interface{})
		for k, v := range m {
			if l.isSensitiveField(k) {
				masked[k] = l.maskValue(v)
			} else {
				masked[k] = v
			}
		}
		return masked
	}

	// Если это struct — конвертируем в map и маскируем
	// (в production лучше использовать reflection аккуратно)
	b, err := json.Marshal(data)
	if err != nil {
		return data
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return data
	}

	return l.maskConfidentialFields(m)
}

// isSensitiveField — определяет является ли поле чувствительным
func (l *Logger) isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{
		"password", "password_hash", "secret", "token", "api_key",
		"email", "phone", "passport", "snils", "inn",
		"credit_card", "ssn", "personal_id",
	}

	fieldName = strings.ToLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldName, sensitive) {
			return true
		}
	}
	return false
}

// maskValue — маскирует значение (частично показывает для удобства)
func (l *Logger) maskValue(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		if len(s) == 0 {
			return "***"
		}
		if len(s) <= 4 {
			return "***"
		}
		// Показываем первые 2 и последние 2 символа
		return s[:2] + "***" + s[len(s)-2:]
	}
	return "***"
}

// ============================================================================
// Helpers: Network
// ============================================================================

func (l *Logger) getClientIP(r *http.Request) string {
	// Проверяем заголовки прокси (для Docker/nginx)
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For может содержать несколько IP
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fallback на RemoteAddr
	return r.RemoteAddr
}

// ============================================================================
// Convenience: Wrapper-методы для распространённых сценариев
// ============================================================================

// LogTaskViewed — логирование просмотра задачи
func (l *Logger) LogTaskViewed(ctx context.Context, r *http.Request, user *model.User, task *model.Task) error {
	return l.Log(ctx, r, user,
		model.ActionSelect, "task", &task.ID,
		task.GetAuditLevel(), // Уровень классификации задачи
		nil, task,
		map[string]interface{}{
			"status_id":  task.StatusID,
			"project_id": task.ProjectID,
			"priority":   task.Priority,
		},
	)
}

// LogTaskCreated — логирование создания задачи
func (l *Logger) LogTaskCreated(ctx context.Context, r *http.Request, user *model.User, task *model.Task) error {
	return l.Log(ctx, r, user,
		model.ActionCreate, "task", &task.ID,
		task.GetAuditLevel(),
		nil, task,
		map[string]interface{}{
			"project_id": task.ProjectID,
		},
	)
}

// LogTaskUpdated — логирование обновления задачи
func (l *Logger) LogTaskUpdated(ctx context.Context, r *http.Request, user *model.User, oldTask, newTask *model.Task) error {
	return l.Log(ctx, r, user,
		model.ActionUpdate, "task", &newTask.ID,
		newTask.GetAuditLevel(),
		oldTask, newTask,
		map[string]interface{}{
			"changed_by": user.Email,
		},
	)
}

// LogTaskDeleted — логирование удаления задачи
func (l *Logger) LogTaskDeleted(ctx context.Context, r *http.Request, user *model.User, task *model.Task) error {
	return l.Log(ctx, r, user,
		model.ActionDelete, "task", &task.ID,
		task.GetAuditLevel(),
		task, nil,
		map[string]interface{}{
			"deleted_by": user.Email,
		},
	)
}

// LogStatusChanged — логирование смены статуса (отдельно для аналитики)
func (l *Logger) LogStatusChanged(ctx context.Context, r *http.Request, user *model.User, taskID, oldStatusID, newStatusID uuid.UUID, transitionName string) error {
	return l.Log(ctx, r, user,
		model.ActionStatusChange, "task", &taskID,
		model.ClassificationConfidential, // Смена статуса — confidential
		map[string]interface{}{"status_id": oldStatusID},
		map[string]interface{}{"status_id": newStatusID},
		map[string]interface{}{
			"transition":  transitionName,
			"from_status": oldStatusID,
			"to_status":   newStatusID,
		},
	)
}

// LogUserLogin — логирование входа пользователя
func (l *Logger) LogUserLogin(ctx context.Context, r *http.Request, user *model.User, success bool, errorMessage string) error {
	metadata := map[string]interface{}{
		"success": success,
		"ip":      l.getClientIP(r),
	}
	if errorMessage != "" {
		metadata["error"] = errorMessage
	}

	return l.Log(ctx, r, user,
		model.ActionLogin, "user", &user.ID,
		model.ClassificationConfidential, // Аутентификация — confidential
		nil, metadata,
		metadata,
	)
}

// LogUserLogout — логирование выхода пользователя
func (l *Logger) LogUserLogout(ctx context.Context, r *http.Request, user *model.User) error {
	return l.Log(ctx, r, user,
		model.ActionLogout, "user", &user.ID,
		model.ClassificationInternal, // Логаут — internal
		nil, nil,
		map[string]interface{}{
			"ip": l.getClientIP(r),
		},
	)
}

// LogProjectViewed — логирование просмотра проекта
func (l *Logger) LogProjectViewed(ctx context.Context, r *http.Request, user *model.User, projectID uuid.UUID, projectName string) error {
	return l.Log(ctx, r, user,
		model.ActionSelect, "project", &projectID,
		model.ClassificationInternal, // Проект — internal (может быть confidential)
		nil, map[string]interface{}{"name": projectName},
		map[string]interface{}{
			"project_name": projectName,
		},
	)
}

// LogExport — логирование экспорта данных (всегда логируем!)
func (l *Logger) LogExport(ctx context.Context, r *http.Request, user *model.User, exportType string, count int, filters map[string]interface{}) error {
	return l.Log(ctx, r, user,
		model.ActionExport, "system", nil,
		model.ClassificationConfidential, // Экспорт — confidential
		nil, map[string]interface{}{
			"export_type": exportType,
			"count":       count,
			"filters":     filters,
		},
		map[string]interface{}{
			"export_type":  exportType,
			"record_count": count,
		},
	)
}

// LogBulkSelect — логирование массового чтения (списки задач)
func (l *Logger) LogBulkSelect(ctx context.Context, r *http.Request, user *model.User, targetType string, count int, filters map[string]interface{}) error {
	// Не логируем если меньше порога
	if count < l.config.SelectBatchThreshold {
		return nil
	}

	return l.Log(ctx, r, user,
		model.ActionSelect, targetType, nil,
		model.ClassificationInternal,
		nil, nil,
		map[string]interface{}{
			"count":   count,
			"bulk":    true,
			"filters": filters,
		},
	)
}

// ============================================================================
// Query: Методы для получения аудита (админка)
// ============================================================================

// GetAuditLogByTarget — получить аудит по сущности (например, история задачи)
func (l *Logger) GetAuditLogByTarget(ctx context.Context, targetType string, targetID uuid.UUID, limit int) ([]*model.AuditLog, error) {
	return l.repo.GetByTarget(ctx, targetType, targetID, limit)
}

// GetAuditLogByUser — получить аудит по пользователю (активность)
func (l *Logger) GetAuditLogByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*model.AuditLog, error) {
	return l.repo.GetByActor(ctx, userID, limit)
}

// GetAuditLogWithFilters — получить аудит с фильтрами (админка)
func (l *Logger) GetAuditLogWithFilters(ctx context.Context, filters *model.AuditFilters) ([]*model.AuditLog, error) {
	return l.repo.GetWithFilters(ctx, filters)
}

// ============================================================================
// Types: Вспомогательные типы
// ============================================================================
