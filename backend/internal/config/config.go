package config

import (
	"fmt"
	"os"
	"strings"
	"time"
	"tracker/internal/model"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// ============================================================================
// Config — корневая структура конфигурации
// ============================================================================
//
// Правила тегирования:
// • `json:"-"` — поле НИКОГДА не должно попадать в JSON (секреты, пароли)
// • `json:"field_name"` — поле можно отдавать в админке/дебаге (безопасно)
// • `mapstructure:"..."` — обязательно для чтения через Viper (YAML/env)
// • Оба тега: `json:"name" mapstructure:"name"` — максимальная гибкость
//
// ============================================================================
type Config struct {
	Database       DatabaseConfig       `json:"database" mapstructure:"database"`
	Server         ServerConfig         `json:"server" mapstructure:"server"`
	Auth           AuthConfig           `json:"auth" mapstructure:"auth"`
	Logging        LoggingConfig        `json:"logging" mapstructure:"logging"`
	Audit          AuditConfig          `json:"audit" mapstructure:"audit"`
	Classification ClassificationConfig `json:"-" mapstructure:"classification"`
}

// ============================================================================
// DatabaseConfig — настройки подключения к БД
// ============================================================================
type DatabaseConfig struct {
	// Driver: sqlite | postgres | mysql
	Driver string `json:"driver" mapstructure:"driver"`

	// Name: для SQLite — путь к файлу (tracker.db), для PostgreSQL — имя БД
	Name string `json:"name" mapstructure:"name"`

	// Host/Port: только для PostgreSQL/MySQL
	Host string `json:"host,omitempty" mapstructure:"host"`
	Port int    `json:"port,omitempty" mapstructure:"port"`

	// Credentials
	User     string `json:"user,omitempty" mapstructure:"user"`
	Password string `json:"-" mapstructure:"password"`

	// SSL/TLS
	SSLMode string `json:"sslmode,omitempty" mapstructure:"sslmode"`

	// Pool settings
	MaxOpenConns    int           `json:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`

	// Вычисляется при загрузке — не читаем из конфига
	DSN string `json:"-" mapstructure:"-"`
}

// ============================================================================
// ServerConfig — настройки HTTP-сервера
// ============================================================================
type ServerConfig struct {
	// Порт сервера (для Docker: 8080)
	Port int `json:"port" mapstructure:"port"`

	// Таймауты — защита от медленных клиентов / DoS
	ReadTimeout     time.Duration `json:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout" mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout" mapstructure:"shutdown_timeout"`
}

// ============================================================================
// AuthConfig — настройки аутентификации и JWT
// ============================================================================
//
// КРИТИЧНО: Все секретные поля имеют json:"-" чтобы не утечь в:
// • Логи (если логируешь конфиг)
// • Ответы админки (GET /api/config)
// • Ошибки валидации
// ============================================================================
type AuthConfig struct {
	// JWT settings
	JWTAlgorithm     string        `json:"jwt_algorithm" mapstructure:"jwt_algorithm"`           // HS256 | ES256 | RS256
	JWTSecret        string        `json:"-" mapstructure:"jwt_secret"`                          // Секрет для HS256
	JWTPrivateKey    string        `json:"-" mapstructure:"jwt_private_key"`                     // Приватный ключ для ES/RS
	JWTPublicKey     string        `json:"-" mapstructure:"jwt_public_key"`                      // Публичный ключ (можно отдавать, но лучше скрыть)
	JWTKeyID         string        `json:"jwt_key_id" mapstructure:"jwt_key_id"`                 // ID ключа (для ротации)
	JWTExpiry        time.Duration `json:"jwt_expiry" mapstructure:"jwt_expiry"`                 // Срок жизни access токена
	JWTRefreshExpiry time.Duration `json:"jwt_refresh_expiry" mapstructure:"jwt_refresh_expiry"` // Срок жизни refresh токена

	// Cookie settings
	CookieSecure      bool   `json:"cookie_secure" mapstructure:"cookie_secure"`             // true для HTTPS
	CookieNameAccess  string `json:"cookie_name_access" mapstructure:"cookie_name_access"`   // Имя куки access токена
	CookieNameRefresh string `json:"cookie_name_refresh" mapstructure:"cookie_name_refresh"` // Имя куки refresh токена
	CookieDomain      string `json:"cookie_domain,omitempty" mapstructure:"cookie_domain"`   // Domain для куки (опционально)
	CookiePath        string `json:"cookie_path" mapstructure:"cookie_path"`                 // Path для куки (обычно "/")
}

// ============================================================================
// LoggingConfig — настройки логирования
// ============================================================================
type LoggingConfig struct {
	// Уровень логов: debug | info | warn | error
	Level string `json:"level" mapstructure:"level"`

	// Формат: text (человекочитаемый) | json (для ELK/Loki)
	Format string `json:"format" mapstructure:"format"`
}

// ============================================================================
// AuditConfig — настройки системы аудита
// ============================================================================
//
// Эти настройки МОЖНО отдавать в админке — они не содержат секретов,
// но показывают как настроена система логирования.
// ============================================================================
type AuditConfig struct {
	// Глобальный переключатель
	Enabled bool `json:"enabled" mapstructure:"enabled"`

	// Какие действия логировать
	LogSelect bool `json:"log_select" mapstructure:"log_select"` // Чтение (просмотр задач)
	LogUpdate bool `json:"log_update" mapstructure:"log_update"` // Изменение
	LogDelete bool `json:"log_delete" mapstructure:"log_delete"` // Удаление
	LogExport bool `json:"log_export" mapstructure:"log_export"` // Экспорт данных

	// Пороги для массовых операций (чтобы не спамить логами)
	SelectBatchThreshold int `json:"select_batch_threshold" mapstructure:"select_batch_threshold"`

	// Хранение логов
	RetentionDays int `json:"retention_days" mapstructure:"retention_days"` // Дней хранения
	MaxRows       int `json:"max_rows" mapstructure:"max_rows"`             // Макс. записей в таблице

	// Селективное логирование
	SelectTargets    []string `json:"select_targets" mapstructure:"select_targets"`         // ["task", "project"] — что логировать при select
	SelectOnlyTagged bool     `json:"select_only_tagged" mapstructure:"select_only_tagged"` // Только задачи с тегом "sensitive"
}

// ============================================================================
// ClassificationConfig — классификация данных и политики аудита
// ============================================================================
//
// json:"-" — реестр полей содержит метаданные о чувствительных данных,
// не должен быть публичным (может помочь атакующему найти слабые места).
// Политики (policies) можно отдавать — они описывают поведение, не данные.
// ============================================================================
type ClassificationConfig struct {
	// Политики аудита по уровням классификации
	// Можно отдавать в админке — показывает как система обрабатывает данные
	Policies map[model.DataClassification]ClassificationPolicy `json:"policies" mapstructure:"policies"`

	// Реестр полей с их классификацией
	// Не отдаём наружу — внутренняя документация системы
	Fields map[string]map[string]DataFieldConfig `json:"-" mapstructure:"fields"`
}

// ClassificationPolicy — правила обработки для уровня классификации
type ClassificationPolicy struct {
	Level            model.DataClassification `json:"level"`
	LogSelect        bool                     `json:"log_select"`
	LogUpdate        bool                     `json:"log_update"`
	LogDelete        bool                     `json:"log_delete"`
	LogExport        bool                     `json:"log_export"`
	RetentionDays    int                      `json:"retention_days"`
	RequiresApproval bool                     `json:"requires_approval"`
	EncryptAtRest    bool                     `json:"encrypt_at_rest"`
	MaskInLogs       bool                     `json:"mask_in_logs"`
}

// DataFieldConfig — описание поля в реестре классификации
type DataFieldConfig struct {
	Type           string                   `json:"type"`           // personal, business, secret
	Classification model.DataClassification `json:"classification"` // public, internal, etc.
	Description    string                   `json:"description"`
}

// ============================================================================
// Load — загрузка конфигурации
// ============================================================================
func Load() (*Config, error) {
	// Загружаем .env (ошибку игнорируем — в prod может не быть)
	_ = godotenv.Load()

	// Настраиваем Viper для основного конфига
	viper.SetConfigFile("configs/config.yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // AUTH_JWT_SECRET → AUTH_JWT_SECRET

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Загружаем отдельные конфиги (аудит, классификация)
	auditConfig, err := loadAuditConfig("configs/audit_policy.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load audit config: %w", err)
	}

	classificationConfig, err := loadClassificationConfig("configs/data_classification.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load classification config: %w", err)
	}

	// Собираем основной конфиг
	cfg := &Config{
		Database: DatabaseConfig{
			Driver:          viper.GetString("database.driver"),
			Name:            viper.GetString("database.name"),
			Host:            viper.GetString("database.host"),
			Port:            viper.GetInt("database.port"), // 👈 Исправлено: был GetString
			User:            viper.GetString("database.user"),
			Password:        getEnv("DB_PASSWORD", viper.GetString("database.password")),
			SSLMode:         viper.GetString("database.sslmode"),
			MaxOpenConns:    viper.GetInt("database.max_open_conns"),
			MaxIdleConns:    viper.GetInt("database.max_idle_conns"),
			ConnMaxLifetime: viper.GetDuration("database.conn_max_lifetime"),
		},
		Server: ServerConfig{
			Port:            viper.GetInt("server.port"), // 👈 Исправлено: был GetString
			ReadTimeout:     viper.GetDuration("server.read_timeout"),
			WriteTimeout:    viper.GetDuration("server.write_timeout"),
			ShutdownTimeout: viper.GetDuration("server.shutdown_timeout"),
		},
		Auth: AuthConfig{
			JWTAlgorithm:      getEnv("JWT_ALGORITHM", viper.GetString("auth.jwt_algorithm")),
			JWTSecret:         getEnv("JWT_SECRET", viper.GetString("auth.jwt_secret")),
			JWTPrivateKey:     getEnv("JWT_PRIVATE_KEY", viper.GetString("auth.jwt_private_key")),
			JWTPublicKey:      getEnv("JWT_PUBLIC_KEY", viper.GetString("auth.jwt_public_key")),
			JWTKeyID:          viper.GetString("auth.jwt_key_id"), // 👈 Исправлено: был jet_key_id
			JWTExpiry:         viper.GetDuration("auth.jwt_expiry"),
			JWTRefreshExpiry:  viper.GetDuration("auth.jwt_refresh_expiry"),
			CookieSecure:      viper.GetBool("auth.cookie_secure"),
			CookieNameAccess:  viper.GetString("auth.cookie_name_access"),
			CookieNameRefresh: viper.GetString("auth.cookie_name_refresh"),
			CookieDomain:      viper.GetString("auth.cookie_domain"),
			CookiePath:        viper.GetString("auth.cookie_path"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("logging.level"),
			Format: viper.GetString("logging.format"),
		},
		Audit:          auditConfig,
		Classification: classificationConfig,
	}

	// Строим DSN для БД
	cfg.Database.DSN = buildDSN(cfg.Database)

	// Валидация
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v\n\n Hint: Copy configs/config.example.yaml to configs/config.yaml", err))
	}

	if err := validate(cfg); err != nil {
		panic(fmt.Sprintf("invalid configuration: %v\n\n💡 Hint: Run './scripts/generate-keys.sh hs256 env > .env.local'", err))
	}

	return cfg
}

// ============================================================================
// Вспомогательные функции
// ============================================================================

// loadAuditConfig — загружает настройки аудита из отдельного YAML
func loadAuditConfig(path string) (AuditConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		// Если файла нет — возвращаем безопасные дефолты
		return getDefaultAuditConfig(), nil
	}

	cfg := AuditConfig{
		Enabled:              v.GetBool("enabled"),
		LogSelect:            v.GetBool("log_select"),
		LogUpdate:            v.GetBool("log_update"),
		LogDelete:            v.GetBool("log_delete"),
		LogExport:            v.GetBool("log_export"),
		SelectBatchThreshold: v.GetInt("select_batch_threshold"),
		RetentionDays:        v.GetInt("retention_days"),
		MaxRows:              v.GetInt("max_rows"),
		SelectTargets:        v.GetStringSlice("select_targets"),
		SelectOnlyTagged:     v.GetBool("select_only_tagged"),
	}

	// Дефолты если не задано
	if !v.IsSet("enabled") {
		cfg.Enabled = true
	}
	if cfg.RetentionDays == 0 {
		cfg.RetentionDays = 90
	}
	if cfg.MaxRows == 0 {
		cfg.MaxRows = 1_000_000
	}
	if cfg.SelectBatchThreshold == 0 {
		cfg.SelectBatchThreshold = 100
	}

	return cfg, nil
}

// loadClassificationConfig — загружает реестр классификации
func loadClassificationConfig(path string) (ClassificationConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		// Если файла нет — возвращаем дефолтные политики
		return ClassificationConfig{
			Policies: getDefaultClassificationPolicies(),
			Fields:   make(map[string]map[string]DataFieldConfig),
		}, nil
	}

	// Парсим политики
	policies := make(map[model.DataClassification]ClassificationPolicy)
	for levelStr, policyData := range v.GetStringMap("policies") {
		level := model.DataClassification(levelStr)
		policy := ClassificationPolicy{Level: level}

		if m, ok := policyData.(map[string]interface{}); ok {
			policy.LogSelect = getBool(m, "log_select", false)
			policy.LogUpdate = getBool(m, "log_update", true)
			policy.LogDelete = getBool(m, "log_delete", true)
			policy.LogExport = getBool(m, "log_export", true)
			policy.RetentionDays = getInt(m, "retention_days", 90)
			policy.RequiresApproval = getBool(m, "requires_approval", false)
			policy.EncryptAtRest = getBool(m, "encrypt_at_rest", false)
			policy.MaskInLogs = getBool(m, "mask_in_logs", false)
		}
		policies[level] = policy
	}

	// Парсим реестр полей
	fields := make(map[string]map[string]DataFieldConfig)
	for entityType, entityData := range v.GetStringMap("fields") {
		if entityMap, ok := entityData.(map[string]interface{}); ok {
			fields[entityType] = make(map[string]DataFieldConfig)
			for fieldName, fieldData := range entityMap {
				if fieldMap, ok := fieldData.(map[string]interface{}); ok {
					fields[entityType][fieldName] = DataFieldConfig{
						Type:           getString(fieldMap, "type", "business"),
						Classification: model.DataClassification(getString(fieldMap, "classification", "internal")),
						Description:    getString(fieldMap, "description", ""),
					}
				}
			}
		}
	}

	return ClassificationConfig{
		Policies: policies,
		Fields:   fields,
	}, nil
}

// getDefaultAuditConfig — безопасные дефолты для аудита
func getDefaultAuditConfig() AuditConfig {
	return AuditConfig{
		Enabled:              true,
		LogSelect:            false, // По умолчанию не логируем чтение (производительность)
		LogUpdate:            true,
		LogDelete:            true,
		LogExport:            true, // Экспорт всегда логируем
		SelectBatchThreshold: 100,  // Не логируем массовое чтение <100 записей
		RetentionDays:        90,
		MaxRows:              1_000_000,
		SelectTargets:        []string{"task", "project"},
		SelectOnlyTagged:     false,
	}
}

// getDefaultClassificationPolicies — дефолтные политики по уровням
func getDefaultClassificationPolicies() map[model.DataClassification]ClassificationPolicy {
	return map[model.DataClassification]ClassificationPolicy{
		model.ClassificationPublic: {
			Level:            model.ClassificationPublic,
			LogSelect:        false,
			LogUpdate:        true,
			LogDelete:        true,
			LogExport:        false,
			RetentionDays:    30,
			RequiresApproval: false,
			EncryptAtRest:    false,
			MaskInLogs:       false,
		},
		model.ClassificationInternal: {
			Level:            model.ClassificationInternal,
			LogSelect:        false,
			LogUpdate:        true,
			LogDelete:        true,
			LogExport:        true,
			RetentionDays:    90,
			RequiresApproval: false,
			EncryptAtRest:    false,
			MaskInLogs:       false,
		},
		model.ClassificationConfidential: {
			Level:            model.ClassificationConfidential,
			LogSelect:        true, // Чтение логируем для ПДн
			LogUpdate:        true,
			LogDelete:        true,
			LogExport:        true,
			RetentionDays:    365,
			RequiresApproval: false,
			EncryptAtRest:    true, // Шифруем в БД
			MaskInLogs:       true, // Маскируем в логах
		},
		model.ClassificationRestricted: {
			Level:            model.ClassificationRestricted,
			LogSelect:        true, // Всегда логируем
			LogUpdate:        true,
			LogDelete:        true,
			LogExport:        true,
			RetentionDays:    2555, // 7 лет (комплаенс)
			RequiresApproval: true, // Требуется одобрение для доступа
			EncryptAtRest:    true,
			MaskInLogs:       true,
		},
	}
}

// Хелперы для безопасного парсинга map
func getBool(m map[string]interface{}, key string, def bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getInt(m map[string]interface{}, key string, def int) int {
	if v, ok := m[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return def
}

func getString(m map[string]interface{}, key string, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

// getEnv — получает значение из env или дефолт из конфига
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// buildDSN — строит connection string для БД
func buildDSN(cfg DatabaseConfig) string {
	switch cfg.Driver {
	case "sqlite":
		return cfg.Name + "?_foreign_keys=on"
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	default:
		return ""
	}
}

// validate — валидация конфигурации
func validate(cfg *Config) error {
	// Database
	if cfg.Database.Driver == "" {
		return fmt.Errorf("database.driver is required")
	}
	if cfg.Database.Driver == "sqlite" && cfg.Database.Name == "" {
		return fmt.Errorf("database.name is required for SQLite")
	}
	if cfg.Database.Driver != "sqlite" {
		if cfg.Database.Host == "" {
			return fmt.Errorf("database.host is required for %s", cfg.Database.Driver)
		}
		if cfg.Database.User == "" {
			return fmt.Errorf("database.user is required for %s", cfg.Database.Driver)
		}
	}

	// Server
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	// Auth
	if cfg.Auth.JWTAlgorithm == "" {
		return fmt.Errorf("auth.jwt_algorithm is required")
	}
	if cfg.Auth.JWTAlgorithm == "HS256" && cfg.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required for HS256")
	}
	if cfg.Auth.JWTExpiry <= 0 {
		return fmt.Errorf("auth.jwt_expiry must be positive")
	}

	// Logging
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info" // Дефолт
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "text" // Дефолт
	}

	// Audit
	if cfg.Audit.RetentionDays < 0 {
		return fmt.Errorf("audit.retention_days cannot be negative")
	}

	return nil
}
