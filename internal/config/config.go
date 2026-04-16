package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Auth     AuthConfig
	Logging  LoggingConfig
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Name            string        `mapstructure:"name"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	DSN             string        `mapstructure:"-"`
}

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type AuthConfig struct {
	JWTSecret    string        `mapstructure:"jwt_secret"`
	JWTAlgorithm string        `mapstructure:"jwt_algorithm"`
	JWTExpiry    time.Duration `mapstructure:"jwt_expiry"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() (*Config, error) {
	//Ignore error because in real prod we haven`t got .env
	_ = godotenv.Load()

	viper.SetConfigFile("configs/config.yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Driver:          viper.GetString("database.driver"),
			Name:            viper.GetString("database.name"),
			Host:            viper.GetString("database.host"),
			Port:            viper.GetInt("database.port"),
			User:            viper.GetString("database.user"),
			Password:        getEnv("DB_PASSWORD", viper.GetString("database.password")),
			SSLMode:         viper.GetString("database.sslmode"),
			MaxOpenConns:    viper.GetInt("database.max_open_conns"),
			MaxIdleConns:    viper.GetInt("database.max_idle_conns"),
			ConnMaxLifetime: viper.GetDuration("database.conn_max_lifetime"),
		},
		Server: ServerConfig{
			Port:            viper.GetString("server.port"),
			ReadTimeout:     viper.GetDuration("server.read_timeout"),
			WriteTimeout:    viper.GetDuration("server.write_timeout"),
			ShutdownTimeout: viper.GetDuration("server.shutdown_timeout"),
		},
		Auth: AuthConfig{ //TODO Update struct to use ES and RS algo
			JWTSecret:    getEnv("JWT_SECRET", viper.GetString("auth.jwt_secret")),
			JWTAlgorithm: getEnv("JWT_ALGORITHM", viper.GetString("auth.jwt_algorithm")),
			JWTExpiry:    viper.GetDuration("auth.jwt_expiry"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("logging.level"),
			Format: viper.GetString("logging.format"),
		},
	}

	cfg.Database.DSN = buildDSN(cfg.Database)

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	if err := validate(cfg); err != nil {
		panic(fmt.Sprintf("invalid configuration: %v\n\n💡 Hint: Run './scripts/generate-keys.sh hs256 env > .env.local'", err))
	}

	return cfg
}

func buildDSN(db DatabaseConfig) string {
	switch db.Driver {
	case "sqlite":
		if db.Name == "" {
			return "file:tracker.db?_foreign_keys=on"
		}
		return fmt.Sprintf("file:%s.db?_foreign_keys=on", db.Name)

	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode,
		)
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true",
			db.User, db.Password, db.Host, db.Port, db.Name,
		)

	default:
		return ""
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func validate(cfg *Config) error {
	if cfg.Database.Driver == "" {
		return fmt.Errorf("database.driver is required")
	}

	if cfg.Auth.JWTSecret == "" || len(cfg.Auth.JWTSecret) < JWT_SECRET_LEN {
		return fmt.Errorf("auth.jwt_secret is required — generate with: ./scripts/generate-keys.sh")
	}

	validAlgorithms := []string{"HS256", "ES256", "RS256"}
	if cfg.Auth.JWTAlgorithm == "" {
		cfg.Auth.JWTAlgorithm = "HS256"
	}
	found := false
	for _, alg := range validAlgorithms {
		if strings.ToUpper(cfg.Auth.JWTAlgorithm) == alg {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unsupported JWT algorithm: %s (supported: HS256, ES256, RS256)", cfg.Auth.JWTAlgorithm)
	}

	if cfg.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}
	if !strings.Contains(cfg.Server.Port, ":") {
		cfg.Server.Port = ":" + cfg.Server.Port
	}

	if cfg.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server.read_timeout must be positive")
	}
	if cfg.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server.write_timeout must be positive")
	}

	return nil
}
