# Конфигурация приложения

## 📊 Архитектура загрузки конфигурации

┌─────────────────────────────────────────────────────────────────┐
│ Data Flow Diagram │
├─────────────────────────────────────────────────────────────────┤
│ │
│ ИСТОЧНИКИ БИБЛИОТЕКИ СТРУКТУРЫ GO │
│ ───────── ────────── ────────────── │
│ │
│ config.yaml ──┐ │
│ │ │
│ .env ────┼──▶ Viper ──▶ map[string]interface{} ──┐ │
│ │ │ │
│ Flags ────┘ │ │
│ │ │
│ encoding/json ───────────────────────┼──▶ struct с json тегами
│ │ │
│ mapstructure ────────────────────────┘ │
│ │
│ ───────────────────────────────────────────────────────────── │
│ КЛЮЧЕВОЙ МОМЕНТ: │
│ • Viper всегда использует mapstructure (не json!) │
│ • encoding/json используется только для Marshal/Unmarshal │
│ • Теги НЕ взаимозаменяемы! │
│ ───────────────────────────────────────────────────────────── │
└─────────────────────────────────────────────────────────────────┘

---

## 🏷️ Правила тегирования полей

### Таблица выбора тегов

| Сценарий | Какой тег | Пример |
|----------|-----------|--------|
| **Конфиг из YAML/.env** | `mapstructure:"..."` | `mapstructure:"jwt_secret"` |
| **Конфиг + JSON API** | `json:"..." mapstructure:"..."` | `json:"port" mapstructure:"port"` |
| **Секреты (не логировать)** | `json:"-" mapstructure:"..."` | `json:"-" mapstructure:"password"` |
| **Только JSON API** | `json:"..."` | `json:"id"` |
| **Вычисляемое поле** | `json:"-" mapstructure:"-"` | `json:"-" mapstructure:"-"` |

---

### Правильные примеры

```go
// Конфиг, который читается из YAML и отдаётся в админку
type ServerConfig struct {
    Port         int           `json:"port" mapstructure:"port"`
    ReadTimeout  time.Duration `json:"read_timeout" mapstructure:"read_timeout"`
}

// Секреты (не должны попасть в JSON-вывод)
type AuthConfig struct {
    JWTSecret     string `json:"-" mapstructure:"jwt_secret"`
    DBPassword    string `json:"-" mapstructure:"password"`
}

// Вычисляемое поле (не читается из конфига)
type DatabaseConfig struct {
    DSN string `json:"-" mapstructure:"-"` // Строится из других полей
}

---

### Неправильные примеры

// Только json тег — Viper НЕ прочитает это из YAML/.env!
type Config struct {
    Secret string `json:"secret"`  // Не сработает с Viper!
}

// Опечатка в теге — поле останется пустым
type Config struct {
    Secret string `mapstructure:"secrect"`  // Опечатка!
}

// Секрет без json:"-" — утечёт в логи/админку
type Config struct {
    Password string `mapstructure:"password"`  // Попадёт в JSON!
}

---

## Что можно отдавать наружу

| Наименование | Можно отдать наружу | Описание |
|--------------|---------------------|----------|
|Server.Port|YES|Номер порта|
|Server.ReadTimeout|YES|Таймауты|
|Logging.Level|YES|Уровень логов|
|Audit.Enabled|YES|Настройки аудита|
|Database.Driver|YES|Тип БД|
|Database.Host|YES|Хост|
|Auth.JWTSecret|NO|Главный секрет JWT
|Auth.JWTPrivateKey|NO|Приватный ключ
|Database.Password|NO|Пароль БД
|Databse.DSN|NO|Содержит пароль
|Classification.Fields|NO|Реестр чувствительных полей

