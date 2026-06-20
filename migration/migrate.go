package migration

import (
	"fmt"
	"log"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	oldmodels "github.com/chennqqi/godnslog/models"
	"xorm.io/xorm"
)

// Migrator handles data migration from 1.0 to 2.0
type Migrator struct {
	engine *xorm.Engine
}

// NewMigrator creates a new migrator
func NewMigrator(engine *xorm.Engine) *Migrator {
	return &Migrator{engine: engine}
}

// MigrateDNS migrates TblDns records to Interaction
func (m *Migrator) MigrateDNS(batchSize int) error {
	log.Println("Starting DNS migration...")

	offset := 0
	totalMigrated := 0

	for {
		// Fetch batch of TblDns records
		var dnsRecords []oldmodels.TblDns
		err := m.engine.Limit(batchSize, offset).Find(&dnsRecords)
		if err != nil {
			return fmt.Errorf("failed to fetch DNS records: %w", err)
		}

		if len(dnsRecords) == 0 {
			break
		}

		// Convert to Interactions
		interactions := make([]*models.Interaction, 0, len(dnsRecords))
		for _, dns := range dnsRecords {
			interaction := models.FromTblDns(&dns)
			interactions = append(interactions, interaction)
		}

		// Insert interactions
		_, err = m.engine.Insert(&interactions)
		if err != nil {
			return fmt.Errorf("failed to insert interactions: %w", err)
		}

		totalMigrated += len(dnsRecords)
		offset += batchSize
		log.Printf("Migrated %d DNS records (total: %d)", len(dnsRecords), totalMigrated)
	}

	log.Printf("DNS migration completed. Total migrated: %d", totalMigrated)
	return nil
}

// MigrateHTTP migrates TblHttp records to Interaction
func (m *Migrator) MigrateHTTP(batchSize int) error {
	log.Println("Starting HTTP migration...")

	offset := 0
	totalMigrated := 0

	for {
		// Fetch batch of TblHttp records
		var httpRecords []oldmodels.TblHttp
		err := m.engine.Limit(batchSize, offset).Find(&httpRecords)
		if err != nil {
			return fmt.Errorf("failed to fetch HTTP records: %w", err)
		}

		if len(httpRecords) == 0 {
			break
		}

		// Convert to Interactions
		interactions := make([]*models.Interaction, 0, len(httpRecords))
		for _, http := range httpRecords {
			interaction := models.FromTblHttp(&http)
			interactions = append(interactions, interaction)
		}

		// Insert interactions
		_, err = m.engine.Insert(&interactions)
		if err != nil {
			return fmt.Errorf("failed to insert interactions: %w", err)
		}

		totalMigrated += len(httpRecords)
		offset += batchSize
		log.Printf("Migrated %d HTTP records (total: %d)", len(httpRecords), totalMigrated)
	}

	log.Printf("HTTP migration completed. Total migrated: %d", totalMigrated)
	return nil
}

// MigrateAll performs all migrations
func (m *Migrator) MigrateAll() error {
	startTime := time.Now()
	log.Println("Starting full migration...")

	// Step 1: Migrate DNS records
	if err := m.MigrateDNS(1000); err != nil {
		return fmt.Errorf("DNS migration failed: %w", err)
	}

	// Step 2: Migrate HTTP records
	if err := m.MigrateHTTP(1000); err != nil {
		return fmt.Errorf("HTTP migration failed: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Full migration completed in %v", elapsed)
	return nil
}

// MigrationStats holds statistics from a migration run
type MigrationStats struct {
	DNSCount     int64
	HTTPCount    int64
	DNSMigrated  int64
	HTTPMigrated int64
	DNSSkipped   int64
	HTTPSkipped  int64
}

// MigrateDNSWithFlags migrates TblDns records to Interaction with dry-run and idempotency support.
func (m *Migrator) MigrateDNSWithFlags(batchSize int, dryRun bool) (*MigrationStats, error) {
	stats := &MigrationStats{}

	dnsCount, err := m.engine.Count(&oldmodels.TblDns{})
	if err != nil {
		return nil, fmt.Errorf("failed to count TblDns: %w", err)
	}
	stats.DNSCount = dnsCount

	if dryRun {
		log.Printf("[dry-run] DNS migration: %d records would be migrated", dnsCount)
		stats.DNSMigrated = dnsCount
		return stats, nil
	}

	offset := 0
	for {
		var dnsRecords []oldmodels.TblDns
		err := m.engine.Limit(batchSize, offset).Find(&dnsRecords)
		if err != nil {
			return stats, fmt.Errorf("failed to fetch DNS records: %w", err)
		}
		if len(dnsRecords) == 0 {
			break
		}

		for _, dns := range dnsRecords {
			interaction := models.FromTblDns(&dns)
			if interaction.Token != nil && *interaction.Token != "" {
				tsStr := interaction.Timestamp.UTC().Format("2006-01-02 15:04:05")
				exists, err := m.engine.Where("token = ? AND strftime('%Y-%m-%d %H:%M:%S', timestamp) = ?",
					*interaction.Token, tsStr).Exist(&models.Interaction{})
				if err != nil {
					return stats, fmt.Errorf("failed to check existing interaction: %w", err)
				}
				if exists {
					stats.DNSSkipped++
					continue
				}
			}
			_, err := m.engine.Insert(interaction)
			if err != nil {
				return stats, fmt.Errorf("failed to insert interaction: %w", err)
			}
			stats.DNSMigrated++
		}

		offset += batchSize
		log.Printf("DNS migration progress: %d migrated, %d skipped", stats.DNSMigrated, stats.DNSSkipped)
	}

	log.Printf("DNS migration completed: %d migrated, %d skipped", stats.DNSMigrated, stats.DNSSkipped)
	return stats, nil
}

// MigrateHTTPWithFlags migrates TblHttp records to Interaction with dry-run and idempotency support.
func (m *Migrator) MigrateHTTPWithFlags(batchSize int, dryRun bool) (*MigrationStats, error) {
	stats := &MigrationStats{}

	httpCount, err := m.engine.Count(&oldmodels.TblHttp{})
	if err != nil {
		return nil, fmt.Errorf("failed to count TblHttp: %w", err)
	}
	stats.HTTPCount = httpCount

	if dryRun {
		log.Printf("[dry-run] HTTP migration: %d records would be migrated", httpCount)
		stats.HTTPMigrated = httpCount
		return stats, nil
	}

	offset := 0
	for {
		var httpRecords []oldmodels.TblHttp
		err := m.engine.Limit(batchSize, offset).Find(&httpRecords)
		if err != nil {
			return stats, fmt.Errorf("failed to fetch HTTP records: %w", err)
		}
		if len(httpRecords) == 0 {
			break
		}

		for _, httpRec := range httpRecords {
			interaction := models.FromTblHttp(&httpRec)
			if interaction.Token != nil && *interaction.Token != "" {
				tsStr := interaction.Timestamp.UTC().Format("2006-01-02 15:04:05")
				exists, err := m.engine.Where("token = ? AND strftime('%Y-%m-%d %H:%M:%S', timestamp) = ?",
					*interaction.Token, tsStr).Exist(&models.Interaction{})
				if err != nil {
					return stats, fmt.Errorf("failed to check existing interaction: %w", err)
				}
				if exists {
					stats.HTTPSkipped++
					continue
				}
			}
			_, err := m.engine.Insert(interaction)
			if err != nil {
				return stats, fmt.Errorf("failed to insert interaction: %w", err)
			}
			stats.HTTPMigrated++
		}

		offset += batchSize
		log.Printf("HTTP migration progress: %d migrated, %d skipped", stats.HTTPMigrated, stats.HTTPSkipped)
	}

	log.Printf("HTTP migration completed: %d migrated, %d skipped", stats.HTTPMigrated, stats.HTTPSkipped)
	return stats, nil
}

// RollbackDNS removes migrated DNS interactions
func (m *Migrator) RollbackDNS() error {
	log.Println("Rolling back DNS migration...")

	_, err := m.engine.Where("type = ?", models.InteractionTypeDNS).Delete(&models.Interaction{})
	if err != nil {
		return fmt.Errorf("failed to rollback DNS interactions: %w", err)
	}

	log.Println("DNS rollback completed")
	return nil
}

// RollbackHTTP removes migrated HTTP interactions
func (m *Migrator) RollbackHTTP() error {
	log.Println("Rolling back HTTP migration...")

	_, err := m.engine.Where("type = ?", models.InteractionTypeHTTP).Delete(&models.Interaction{})
	if err != nil {
		return fmt.Errorf("failed to rollback HTTP interactions: %w", err)
	}

	log.Println("HTTP rollback completed")
	return nil
}

// RollbackAll performs full rollback
func (m *Migrator) RollbackAll() error {
	log.Println("Starting full rollback...")

	if err := m.RollbackDNS(); err != nil {
		return err
	}

	if err := m.RollbackHTTP(); err != nil {
		return err
	}

	log.Println("Full rollback completed")
	return nil
}

// ValidateMigration checks if migration was successful
func (m *Migrator) ValidateMigration() error {
	log.Println("Validating migration...")

	// Count TblDns
	dnsCount, err := m.engine.Count(&oldmodels.TblDns{})
	if err != nil {
		return fmt.Errorf("failed to count TblDns: %w", err)
	}

	// Count TblHttp
	httpCount, err := m.engine.Count(&oldmodels.TblHttp{})
	if err != nil {
		return fmt.Errorf("failed to count TblHttp: %w", err)
	}

	// Count DNS interactions
	dnsInteractionCount, err := m.engine.Where("type = ?", models.InteractionTypeDNS).Count(&models.Interaction{})
	if err != nil {
		return fmt.Errorf("failed to count DNS interactions: %w", err)
	}

	// Count HTTP interactions
	httpInteractionCount, err := m.engine.Where("type = ?", models.InteractionTypeHTTP).Count(&models.Interaction{})
	if err != nil {
		return fmt.Errorf("failed to count HTTP interactions: %w", err)
	}

	log.Printf("TblDns count: %d, DNS interactions: %d", dnsCount, dnsInteractionCount)
	log.Printf("TblHttp count: %d, HTTP interactions: %d", httpCount, httpInteractionCount)

	if dnsCount != dnsInteractionCount {
		return fmt.Errorf("DNS count mismatch: TblDns=%d, Interactions=%d", dnsCount, dnsInteractionCount)
	}

	if httpCount != httpInteractionCount {
		return fmt.Errorf("HTTP count mismatch: TblHttp=%d, Interactions=%d", httpCount, httpInteractionCount)
	}

	log.Println("Migration validation passed")
	return nil
}
