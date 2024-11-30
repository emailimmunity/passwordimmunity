package services

import (
	"context"
	"time"
	"database/sql"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"github.com/golang-migrate/migrate/v4"
)

type MigrationService interface {
	RunMigrations(ctx context.Context) error
	RollbackMigration(ctx context.Context, version uint) error
	GetMigrationHistory(ctx context.Context) ([]models.Migration, error)
	CreateBackup(ctx context.Context) error
	ValidateSchema(ctx context.Context) error
}

type migrationService struct {
	repo        repository.Repository
	audit       AuditService
	db          *sql.DB
}

func NewMigrationService(repo repository.Repository, audit AuditService, db *sql.DB) MigrationService {
	return &migrationService{
		repo:    repo,
		audit:   audit,
		db:      db,
	}
}

func (s *migrationService) RunMigrations(ctx context.Context) error {
	// Create backup before migration
	if err := s.CreateBackup(ctx); err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres",
		&migrate.DatabaseConfig{},
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		return err
	}

	migration := &models.Migration{
		ID:        uuid.New(),
		Version:   version,
		Status:    "completed",
		Dirty:     dirty,
		AppliedAt: time.Now(),
	}

	if err := s.repo.CreateMigration(ctx, migration); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("migration_completed", "Database migration completed")
	metadata["version"] = version
	if err := s.createAuditLog(ctx, "migration.completed", uuid.Nil, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *migrationService) RollbackMigration(ctx context.Context, version uint) error {
	// Create backup before rollback
	if err := s.CreateBackup(ctx); err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres",
		&migrate.DatabaseConfig{},
	)
	if err != nil {
		return err
	}

	if err := m.Steps(-1); err != nil {
		return err
	}

	migration := &models.Migration{
		ID:        uuid.New(),
		Version:   version,
		Status:    "rolledback",
		Dirty:     false,
		AppliedAt: time.Now(),
	}

	if err := s.repo.CreateMigration(ctx, migration); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("migration_rolledback", "Database migration rolled back")
	metadata["version"] = version
	if err := s.createAuditLog(ctx, "migration.rolledback", uuid.Nil, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *migrationService) GetMigrationHistory(ctx context.Context) ([]models.Migration, error) {
	return s.repo.ListMigrations(ctx)
}

func (s *migrationService) CreateBackup(ctx context.Context) error {
	// Implement database backup logic
	return nil
}

func (s *migrationService) ValidateSchema(ctx context.Context) error {
	// Implement schema validation logic
	return nil
}
