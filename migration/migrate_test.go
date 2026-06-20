package migration

import (
	"testing"
	"time"

	newmodels "github.com/chennqqi/godnslog/internal/models"
	oldmodels "github.com/chennqqi/godnslog/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupMigrationEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	if err := engine.Sync2(
		new(newmodels.Interaction),
		new(oldmodels.TblDns),
		new(oldmodels.TblHttp),
	); err != nil {
		t.Fatalf("failed to sync tables: %v", err)
	}
	return engine
}

func TestMigrateDNS_DryRun(t *testing.T) {
	engine := setupMigrationEngine(t)
	migrator := NewMigrator(engine)

	ts := time.Now()
	records := []oldmodels.TblDns{
		{Domain: "a.example.com", Var: "tok1", Ip: "1.2.3.4", Ctime: ts, Atime: ts},
		{Domain: "b.example.com", Var: "tok2", Ip: "5.6.7.8", Ctime: ts, Atime: ts},
	}
	for i := range records {
		if _, err := engine.Insert(&records[i]); err != nil {
			t.Fatalf("failed to insert TblDns: %v", err)
		}
	}

	stats, err := migrator.MigrateDNSWithFlags(1000, true)
	if err != nil {
		t.Fatalf("MigrateDNSWithFlags dry-run failed: %v", err)
	}
	if stats.DNSCount != 2 {
		t.Errorf("expected DNSCount=2, got %d", stats.DNSCount)
	}

	// Verify no interactions were inserted
	count, err := engine.Where("type = ?", newmodels.InteractionTypeDNS).Count(&newmodels.Interaction{})
	if err != nil {
		t.Fatalf("failed to count interactions: %v", err)
	}
	if count != 0 {
		t.Errorf("dry-run should not insert any interactions, got %d", count)
	}
}

func TestMigrateDNS_Idempotent(t *testing.T) {
	engine := setupMigrationEngine(t)
	migrator := NewMigrator(engine)

	ts := time.Now()
	records := []oldmodels.TblDns{
		{Domain: "a.example.com", Var: "tok1", Ip: "1.2.3.4", Ctime: ts, Atime: ts},
		{Domain: "b.example.com", Var: "tok2", Ip: "5.6.7.8", Ctime: ts, Atime: ts},
	}
	for i := range records {
		if _, err := engine.Insert(&records[i]); err != nil {
			t.Fatalf("failed to insert TblDns: %v", err)
		}
	}

	// First run — should migrate 2
	stats1, err := migrator.MigrateDNSWithFlags(1000, false)
	if err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if stats1.DNSMigrated != 2 {
		t.Errorf("expected 2 migrated, got %d", stats1.DNSMigrated)
	}

	// Second run — should migrate 0 (idempotent)
	stats2, err := migrator.MigrateDNSWithFlags(1000, false)
	if err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
	if stats2.DNSMigrated != 0 {
		t.Errorf("expected 0 migrated on second run, got %d", stats2.DNSMigrated)
	}
	if stats2.DNSSkipped != 2 {
		t.Errorf("expected 2 skipped on second run, got %d", stats2.DNSSkipped)
	}
}

func TestMigrateHTTP_DryRun(t *testing.T) {
	engine := setupMigrationEngine(t)
	migrator := NewMigrator(engine)

	ts := time.Now()
	records := []oldmodels.TblHttp{
		{Ip: "1.2.3.4", Var: "tok1", Path: "/test", Method: "GET", Ctime: ts, Atime: ts},
		{Ip: "5.6.7.8", Var: "tok2", Path: "/test2", Method: "POST", Ctime: ts, Atime: ts},
	}
	for i := range records {
		if _, err := engine.Insert(&records[i]); err != nil {
			t.Fatalf("failed to insert TblHttp: %v", err)
		}
	}

	stats, err := migrator.MigrateHTTPWithFlags(1000, true)
	if err != nil {
		t.Fatalf("MigrateHTTPWithFlags dry-run failed: %v", err)
	}
	if stats.HTTPCount != 2 {
		t.Errorf("expected HTTPCount=2, got %d", stats.HTTPCount)
	}

	count, err := engine.Where("type = ?", newmodels.InteractionTypeHTTP).Count(&newmodels.Interaction{})
	if err != nil {
		t.Fatalf("failed to count interactions: %v", err)
	}
	if count != 0 {
		t.Errorf("dry-run should not insert any interactions, got %d", count)
	}
}

func TestMigrateStats(t *testing.T) {
	engine := setupMigrationEngine(t)
	migrator := NewMigrator(engine)

	ts := time.Now()
	dnsRecords := []oldmodels.TblDns{
		{Domain: "a.example.com", Var: "tok1", Ip: "1.2.3.4", Ctime: ts, Atime: ts},
	}
	httpRecords := []oldmodels.TblHttp{
		{Ip: "5.6.7.8", Var: "tok2", Path: "/test", Method: "GET", Ctime: ts, Atime: ts},
	}
	for i := range dnsRecords {
		engine.Insert(&dnsRecords[i])
	}
	for i := range httpRecords {
		engine.Insert(&httpRecords[i])
	}

	dnsStats, err := migrator.MigrateDNSWithFlags(1000, false)
	if err != nil {
		t.Fatalf("DNS migration failed: %v", err)
	}
	httpStats, err := migrator.MigrateHTTPWithFlags(1000, false)
	if err != nil {
		t.Fatalf("HTTP migration failed: %v", err)
	}

	if dnsStats.DNSCount != 1 || dnsStats.DNSMigrated != 1 || dnsStats.DNSSkipped != 0 {
		t.Errorf("unexpected DNS stats: %+v", dnsStats)
	}
	if httpStats.HTTPCount != 1 || httpStats.HTTPMigrated != 1 || httpStats.HTTPSkipped != 0 {
		t.Errorf("unexpected HTTP stats: %+v", httpStats)
	}
}

func TestSyncCommand_DryRun(t *testing.T) {
	engine := setupMigrationEngine(t)

	ts := time.Now()
	dnsRecords := []oldmodels.TblDns{
		{Domain: "a.example.com", Var: "tok1", Ip: "1.2.3.4", Ctime: ts, Atime: ts},
	}
	for i := range dnsRecords {
		engine.Insert(&dnsRecords[i])
	}

	stats, err := RunSync(engine, true)
	if err != nil {
		t.Fatalf("RunSync dry-run failed: %v", err)
	}
	if stats.DNSCount != 1 {
		t.Errorf("expected DNSCount=1, got %d", stats.DNSCount)
	}

	// Verify no interactions inserted
	count, _ := engine.Count(&newmodels.Interaction{})
	if count != 0 {
		t.Errorf("dry-run should not insert interactions, got %d", count)
	}
}

func TestSyncCommand_RealRun(t *testing.T) {
	engine := setupMigrationEngine(t)

	ts := time.Now()
	dnsRecords := []oldmodels.TblDns{
		{Domain: "a.example.com", Var: "tok1", Ip: "1.2.3.4", Ctime: ts, Atime: ts},
		{Domain: "b.example.com", Var: "tok2", Ip: "5.6.7.8", Ctime: ts, Atime: ts},
	}
	for i := range dnsRecords {
		engine.Insert(&dnsRecords[i])
	}

	stats, err := RunSync(engine, false)
	if err != nil {
		t.Fatalf("RunSync failed: %v", err)
	}
	if stats.DNSMigrated != 2 {
		t.Errorf("expected 2 migrated, got %d", stats.DNSMigrated)
	}

	count, _ := engine.Count(&newmodels.Interaction{})
	if count != 2 {
		t.Errorf("expected 2 interactions, got %d", count)
	}
}

func TestSyncCommand_Idempotent(t *testing.T) {
	engine := setupMigrationEngine(t)

	ts := time.Now()
	dnsRecords := []oldmodels.TblDns{
		{Domain: "a.example.com", Var: "tok1", Ip: "1.2.3.4", Ctime: ts, Atime: ts},
	}
	for i := range dnsRecords {
		engine.Insert(&dnsRecords[i])
	}

	// First run
	stats1, err := RunSync(engine, false)
	if err != nil {
		t.Fatalf("first RunSync failed: %v", err)
	}
	if stats1.DNSMigrated != 1 {
		t.Errorf("expected 1 migrated, got %d", stats1.DNSMigrated)
	}

	// Second run — should be idempotent
	stats2, err := RunSync(engine, false)
	if err != nil {
		t.Fatalf("second RunSync failed: %v", err)
	}
	if stats2.DNSMigrated != 0 {
		t.Errorf("expected 0 migrated on second run, got %d", stats2.DNSMigrated)
	}
	if stats2.DNSSkipped != 1 {
		t.Errorf("expected 1 skipped on second run, got %d", stats2.DNSSkipped)
	}
}

func TestFromTblDns(t *testing.T) {
	testTime := time.Now()

	dns := &oldmodels.TblDns{
		Id:     1,
		Uid:    100,
		Domain: "test.example.com",
		Var:    "abc123",
		Ip:     "192.168.1.1",
		Ctime:  testTime,
		Atime:  testTime,
	}

	interaction := newmodels.FromTblDns(dns)

	if interaction.Type != newmodels.InteractionTypeDNS {
		t.Errorf("Expected type %s, got %s", newmodels.InteractionTypeDNS, interaction.Type)
	}

	if interaction.Domain == nil || *interaction.Domain != dns.Domain {
		t.Errorf("Expected domain %s, got %v", dns.Domain, interaction.Domain)
	}

	if interaction.Token == nil || *interaction.Token != dns.Var {
		t.Errorf("Expected token %s, got %v", dns.Var, interaction.Token)
	}

	if interaction.SourceIP != dns.Ip {
		t.Errorf("Expected source IP %s, got %s", dns.Ip, interaction.SourceIP)
	}
}

func TestFromTblHttp(t *testing.T) {
	testTime := time.Now()

	http := &oldmodels.TblHttp{
		Id:     1,
		Uid:    100,
		Ip:     "192.168.1.1",
		Var:    "abc123",
		Path:   "/test",
		Method: "GET",
		Data:   "test data",
		Ctype:  "application/json",
		Ua:     "Mozilla/5.0",
		Ctime:  testTime,
		Atime:  testTime,
	}

	interaction := newmodels.FromTblHttp(http)

	if interaction.Type != newmodels.InteractionTypeHTTP {
		t.Errorf("Expected type %s, got %s", newmodels.InteractionTypeHTTP, interaction.Type)
	}

	if interaction.Method == nil || *interaction.Method != http.Method {
		t.Errorf("Expected method %s, got %v", http.Method, interaction.Method)
	}

	if interaction.Path == nil || *interaction.Path != http.Path {
		t.Errorf("Expected path %s, got %v", http.Path, interaction.Path)
	}

	if interaction.SourceIP != http.Ip {
		t.Errorf("Expected source IP %s, got %s", http.Ip, interaction.SourceIP)
	}
}
