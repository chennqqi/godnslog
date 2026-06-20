package migration

import (
	"flag"
	"fmt"
	"log"
	"os"

	"xorm.io/xorm"
)

// RunSync executes the migration with the given options.
// Returns combined MigrationStats from DNS and HTTP migration.
func RunSync(engine *xorm.Engine, dryRun bool) (*MigrationStats, error) {
	migrator := NewMigrator(engine)

	dnsStats, err := migrator.MigrateDNSWithFlags(1000, dryRun)
	if err != nil {
		return nil, fmt.Errorf("DNS migration failed: %w", err)
	}

	httpStats, err := migrator.MigrateHTTPWithFlags(1000, dryRun)
	if err != nil {
		return nil, fmt.Errorf("HTTP migration failed: %w", err)
	}

	combined := &MigrationStats{
		DNSCount:     dnsStats.DNSCount,
		HTTPCount:    httpStats.HTTPCount,
		DNSMigrated:  dnsStats.DNSMigrated,
		HTTPMigrated: httpStats.HTTPMigrated,
		DNSSkipped:   dnsStats.DNSSkipped,
		HTTPSkipped:  httpStats.HTTPSkipped,
	}
	return combined, nil
}

// Main is the CLI entry point for the sync command.
func Main() {
	driver := flag.String("driver", "sqlite", "database driver")
	dsn := flag.String("dsn", "", "database DSN")
	dryRun := flag.Bool("dry-run", false, "dry run mode (no actual writes)")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("dsn is required")
	}

	engine, err := xorm.NewEngine(*driver, *dsn)
	if err != nil {
		log.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	stats, err := RunSync(engine, *dryRun)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	mode := "live"
	if *dryRun {
		mode = "dry-run"
	}
	log.Printf("Migration complete (%s): DNS=%d migrated, %d skipped; HTTP=%d migrated, %d skipped",
		mode, stats.DNSMigrated, stats.DNSSkipped, stats.HTTPMigrated, stats.HTTPSkipped)
	_ = os.Stdout.Sync()
}
