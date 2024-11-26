package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/emailimmunity/passwordimmunity/db"
)

func main() {
	var (
		direction   string
		dbURL      string
		migrations string
	)

	flag.StringVar(&direction, "direction", "up", "Migration direction (up/down)")
	flag.StringVar(&dbURL, "db", os.Getenv("DATABASE_URL"), "Database URL")
	flag.StringVar(&migrations, "dir", "db/migrations", "Migrations directory")
	flag.Parse()

	if dbURL == "" {
		log.Fatal("Database URL is required")
	}

	migrator, err := db.NewMigrator(dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}

	migrationsPath, err := filepath.Abs(migrations)
	if err != nil {
		log.Fatalf("Failed to resolve migrations path: %v", err)
	}

	if err := migrator.LoadMigrations(migrationsPath); err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	switch direction {
	case "up":
		if err := migrator.MigrateUp(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migration completed successfully")
	case "down":
		if err := migrator.MigrateDown(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully")
	default:
		log.Fatalf("Invalid direction: %s", direction)
	}
}
