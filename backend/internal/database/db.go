package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL
	_ "github.com/lib/pq"              // PostgreSQL
	_ "modernc.org/sqlite"             // SQLite

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrationsFS embed.FS

func New(ctx context.Context, driver, dsn string) (*sql.DB, error) {
	if !isValidDriver(driver) {
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Hour)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if err := Migrate(driver, dsn); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	log.Printf("database initialized (driver=%s)", driver)
	return db, nil
}

// Migrate применяет миграции из embed-filesystem
func Migrate(driver, dsn string) error {
	// 1. Путь внутри embed.FS
	migrationsSubpath := "sqlite" // Для SQLite используем универсальные миграции
	// Если нужны отдельные папки для PostgreSQL/MySQL:
	// if driver != "sqlite" {
	//     migrationsSubpath = driver
	// }

	// 2. Создаём источник миграций из embed.FS
	sourceDriver, err := iofs.New(migrationsFS, fmt.Sprintf("migrations/%s", migrationsSubpath))
	if err != nil {
		return fmt.Errorf("failed to init migrate source: %w", err)
	}

	// 3. Формируем DSN для драйвера БД (конвертация форматов)
	migrateDSN, err := buildMigrateDSN(driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to build migrate DSN: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		sourceDriver,
		migrateDSN,
	)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}
	defer m.Close()

	// 5. Применяем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("migrations applied successfully")
	return nil
}

// buildMigrateDSN строит DSN для драйвера БД (не для источника!)
func buildMigrateDSN(driver, appDSN string) (string, error) {
	switch driver {
	case "sqlite":
		// Убираем "file:" префикс для migrate
		dbPath := strings.TrimPrefix(appDSN, "file:")
		return "sqlite://" + dbPath, nil
	case "postgres":
		if !strings.HasPrefix(appDSN, "postgres://") && !strings.HasPrefix(appDSN, "postgresql://") {
			return "postgres://" + appDSN, nil
		}
		return appDSN, nil
	case "mysql":
		if !strings.HasPrefix(appDSN, "mysql://") {
			return "mysql://" + appDSN, nil
		}
		return appDSN, nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", driver)
	}
}

func isValidDriver(driver string) bool {
	validDrivers := []string{"sqlite", "postgres", "mysql"}
	for _, v := range validDrivers {
		if driver == v {
			return true
		}
	}
	return false
}
