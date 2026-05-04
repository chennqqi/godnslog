package retention

import (
	"testing"
	"time"
)

// TestRetentionPolicyModel tests retention policy model
func TestRetentionPolicyModel(t *testing.T) {
	now := time.Now()
	policy := &RetentionPolicy{
		ID:                  "policy-1",
		Name:                "Default Policy",
		Description:         "Default retention policy",
		ApplyToInteractions: true,
		ApplyToCases:        true,
		ApplyToPayloads:     true,
		RetentionDays:       90,
		MaxRecords:          10000,
		ArchiveAfterDays:    30,
		IsEnabled:           true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if policy.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if policy.Name == "" {
		t.Fatal("Name should not be empty")
	}

	if policy.RetentionDays < 0 {
		t.Fatal("RetentionDays should be non-negative")
	}
}

// TestArchiveModel tests archive model
func TestArchiveModel(t *testing.T) {
	now := time.Now()
	archive := &Archive{
		ID:          "archive-1",
		PolicyID:    "policy-1",
		DataType:    "interactions",
		RecordCount: 1000,
		StoragePath: "/archives/archive-1.gz",
		FileSize:    1024000,
		Checksum:    "abc123",
		Compression: "gzip",
		StartTime:   now.AddDate(0, 0, -30),
		EndTime:     now,
		Status:      "completed",
		CreatedAt:   now,
	}

	if archive.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if archive.DataType == "" {
		t.Fatal("DataType should not be empty")
	}

	if archive.StoragePath == "" {
		t.Fatal("StoragePath should not be empty")
	}
}

// TestRetentionJobModel tests retention job model
func TestRetentionJobModel(t *testing.T) {
	now := time.Now()
	job := &RetentionJob{
		ID:               "job-1",
		PolicyID:         "policy-1",
		JobType:          "retention",
		Status:           "completed",
		RecordsProcessed: 1000,
		RecordsDeleted:   500,
		StartedAt:        now.Add(-time.Hour),
		CreatedAt:        now,
	}

	if job.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if job.JobType == "" {
		t.Fatal("JobType should not be empty")
	}

	if job.Status == "" {
		t.Fatal("Status should not be empty")
	}
}

// TestTableName tests table names
func TestTableName(t *testing.T) {
	policy := RetentionPolicy{}
	if policy.TableName() != "retention_policies" {
		t.Fatalf("Expected 'retention_policies', got '%s'", policy.TableName())
	}

	archive := Archive{}
	if archive.TableName() != "archives" {
		t.Fatalf("Expected 'archives', got '%s'", archive.TableName())
	}

	job := RetentionJob{}
	if job.TableName() != "retention_jobs" {
		t.Fatalf("Expected 'retention_jobs', got '%s'", job.TableName())
	}
}
