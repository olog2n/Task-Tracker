package model

type DataClassification string

const (
	ClassificationPublic       DataClassification = "public"       // Без ограничений
	ClassificationInternal     DataClassification = "internal"     // Служебное
	ClassificationConfidential DataClassification = "confidential" // ПДн, коммерческая тайна
	ClassificationRestricted   DataClassification = "restricted"   // Критичные данные
)

// DataType — тип данных (для документации)
type DataType string

const (
	DataTypeTechnical DataType = "technical" // ID, хеши
	DataTypePersonal  DataType = "personal"  // ПДн
	DataTypeBusiness  DataType = "business"  // Бизнес-данные
	DataTypeFinancial DataType = "financial" // Финансы
	DataTypeSecret    DataType = "secret"    // Секреты, ключи
	DataTypeAudit     DataType = "audit"     // Аудит-логи
)

// DataField — описание поля с классификацией
type DataField struct {
	Name           string             `json:"name"`
	DataType       DataType           `json:"data_type"`
	Classification DataClassification `json:"classification"`
	Description    string             `json:"description"`
}

// ClassificationPolicy — политика аудита для уровня
type ClassificationPolicy struct {
	Level            DataClassification `json:"level"`
	LogSelect        bool               `json:"log_select"`        // Логировать чтение
	LogUpdate        bool               `json:"log_update"`        // Логировать изменение
	LogDelete        bool               `json:"log_delete"`        // Логировать удаление
	LogExport        bool               `json:"log_export"`        // Логировать экспорт
	RetentionDays    int                `json:"retention_days"`    // Срок хранения
	RequiresApproval bool               `json:"requires_approval"` // Требует одобрения для доступа
	EncryptAtRest    bool               `json:"encrypt_at_rest"`   // Шифрование в БД
	MaskInLogs       bool               `json:"mask_in_logs"`      // Маскировать в логах
}
