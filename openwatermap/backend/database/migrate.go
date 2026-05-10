package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations запускает все pending миграции
func RunMigrations(conn *sql.DB) error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("ошибка инициализации источника миграций: %w", err)
	}

	// PostgreSQL драйвер для migrate
	dbDriver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("ошибка инициализации драйвера миграций: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("ошибка создания мигратора: %w", err)
	}

	// Применить все новые миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}

	version, dirty, _ := m.Version()
	if dirty {
		log.Printf("⚠️  Миграция %d в грязном состоянии!", version)
	} else {
		log.Printf("✅ Миграции применены, версия: %d", version)
	}

	return nil
}
