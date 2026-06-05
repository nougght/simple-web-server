package postgres

import (
	"errors"
	"fmt"
	"log"
	"simple-server/internal/config"

	"github.com/golang-migrate/migrate/v4"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// запуск миграций
func RunMigrations(db *sqlx.DB) error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	dbDriver, err := pg.WithInstance(db.DB, &pg.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no changes to migrate")
			return nil
		}
		return fmt.Errorf("migrate up failed: %w", err)
	}

	return nil
}

// подключение к БД и запуск миграций
func ConnectDB(config *config.PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}
	log.Println("successful migrations")
	return db, nil
}
