package migrate

import (
    "database/sql"
    "embed"
    "fmt"
    "log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Run executes all pending migrations
func Run(db *sql.DB) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return fmt.Errorf("could not create database driver: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://db/migrations",
        "postgres", driver)
    if err != nil {
        return fmt.Errorf("could not create migrate instance: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("could not run migrations: %v", err)
    }

    version, dirty, err := m.Version()
    if err != nil && err != migrate.ErrNilVersion {
        return fmt.Errorf("could not get migration version: %v", err)
    }

    log.Printf("Database migrated to version %d (dirty: %v)", version, dirty)
    return nil
}

// Rollback reverts the last applied migration
func Rollback(db *sql.DB) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return fmt.Errorf("could not create database driver: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://db/migrations",
        "postgres", driver)
    if err != nil {
        return fmt.Errorf("could not create migrate instance: %v", err)
    }

    if err := m.Steps(-1); err != nil {
        return fmt.Errorf("could not rollback migration: %v", err)
    }

    return nil
}
