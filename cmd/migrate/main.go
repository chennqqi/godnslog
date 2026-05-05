package main

import (
	"flag"
	"log"
	"os"

	"github.com/chennqqi/godnslog/migration"
	"xorm.io/xorm"
)

func main() {
	rollback := flag.Bool("rollback", false, "Rollback migration")
	validate := flag.Bool("validate", false, "Validate migration only")
	datasource := flag.String("db", "", "Database connection string")

	flag.Parse()

	// Get datasource from flag
	ds := *datasource
	if ds == "" {
		log.Fatal("Database connection string is required. Use -db flag")
	}

	// Initialize engine
	engine, err := xorm.NewEngine("sqlite3", ds)
	if err != nil {
		log.Fatalf("Failed to initialize database engine: %v", err)
	}
	defer engine.Close()

	// Create migrator
	migrator := migration.NewMigrator(engine)

	// Execute command
	switch {
	case *rollback:
		log.Println("Rolling back migration...")
		if err := migrator.RollbackAll(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully")

	case *validate:
		log.Println("Validating migration...")
		if err := migrator.ValidateMigration(); err != nil {
			log.Fatalf("Validation failed: %v", err)
			os.Exit(1)
		}
		log.Println("Validation passed")

	default:
		log.Println("Starting migration...")
		if err := migrator.MigrateAll(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migration completed successfully")

		// Validate after migration
		if err := migrator.ValidateMigration(); err != nil {
			log.Fatalf("Validation failed: %v", err)
			os.Exit(1)
		}
		log.Println("Validation passed")
	}
}
