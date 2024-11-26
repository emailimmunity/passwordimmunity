package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// Migration represents a database migration
type Migration struct {
	Version  int
	UpFile   string
	DownFile string
}

// Migrator handles database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator creates a new migrator instance
func NewMigrator(dbURL string) (*Migrator, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &Migrator{db: db}, nil
}

// LoadMigrations loads migration files from the migrations directory
func (m *Migrator) LoadMigrations(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	migrations := make(map[int]Migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".sql") {
			continue
		}

		var version int
		var isRollback bool
		_, err := fmt.Sscanf(filename, "%06d", &version)
		if err != nil {
			continue
		}

		isRollback = strings.Contains(filename, "rollback")
		fullPath := filepath.Join(dir, filename)

		if isRollback {
			if mig, ok := migrations[version]; ok {
				mig.DownFile = fullPath
				migrations[version] = mig
			} else {
				migrations[version] = Migration{
					Version:  version,
					DownFile: fullPath,
				}
			}
		} else {
			if mig, ok := migrations[version]; ok {
				mig.UpFile = fullPath
				migrations[version] = mig
			} else {
				migrations[version] = Migration{
					Version:  version,
					UpFile:   fullPath,
				}
			}
		}
	}

	// Convert map to slice and sort
	m.migrations = make([]Migration, 0, len(migrations))
	for _, migration := range migrations {
		m.migrations = append(m.migrations, migration)
	}
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// MigrateUp runs all pending migrations
func (m *Migrator) MigrateUp() error {
	// Create migrations table if it doesn't exist
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version integer PRIMARY KEY,
			applied_at timestamp with time zone DEFAULT current_timestamp
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	for _, migration := range m.migrations {
		var applied bool
		err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", migration.Version).Scan(&applied)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

		if !applied {
			sql, err := os.ReadFile(migration.UpFile)
			if err != nil {
				return fmt.Errorf("failed to read migration file: %v", err)
			}

			tx, err := m.db.Begin()
			if err != nil {
				return fmt.Errorf("failed to start transaction: %v", err)
			}

			if _, err := tx.Exec(string(sql)); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to apply migration: %v", err)
			}

			if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration: %v", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %v", err)
			}

			log.Printf("Applied migration %d", migration.Version)
		}
	}

	return nil
}

// MigrateDown rolls back the last migration
func (m *Migrator) MigrateDown() error {
	var lastVersion int
	err := m.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&lastVersion)
	if err != nil {
		return fmt.Errorf("failed to get last migration: %v", err)
	}

	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if migration.Version == lastVersion {
			sql, err := os.ReadFile(migration.DownFile)
			if err != nil {
				return fmt.Errorf("failed to read rollback file: %v", err)
			}

			tx, err := m.db.Begin()
			if err != nil {
				return fmt.Errorf("failed to start transaction: %v", err)
			}

			if _, err := tx.Exec(string(sql)); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to rollback migration: %v", err)
			}


			if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", migration.Version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record rollback: %v", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %v", err)
			}

			log.Printf("Rolled back migration %d", migration.Version)
			break
		}
	}

	return nil
}
