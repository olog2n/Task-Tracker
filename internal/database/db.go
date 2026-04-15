package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	// 👇 Импортируем ВСЕ драйверы (side-effect через _)
	_ "github.com/go-sql-driver/mysql" // MySQL
	_ "github.com/lib/pq"              // PostgreSQL
	_ "modernc.org/sqlite"             // SQLite

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

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

func Migrate(driver, dsn string) error {
	migrationsPath := fmt.Sprintf("file://internal/database/migrations/%s", driver)

	migrateDSN, err := buildMigrateDSN(driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to build migrate DSN: %w", err)
	}

	m, err := migrate.New(migrationsPath, migrateDSN)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("migrations applied successfully")
	return nil
}

func buildMigrateDSN(driver, dsn string) (string, error) {
	switch driver {
	case "sqlite":
		if !strings.HasPrefix(dsn, "sqlite://") {
			return "sqlite://" + dsn, nil
		}
		return dsn, nil

	case "postgres":
		if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
			return "postgres://" + dsn, nil
		}
		return dsn, nil

	case "mysql":
		if !strings.HasPrefix(dsn, "mysql://") {
			return "mysql://" + dsn, nil
		}
		return dsn, nil

	default:
		return "", fmt.Errorf("unsupported driver for migrations: %s", driver)
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
