package retention

import "time"

// RetentionPolicy defines how long data should be kept
type RetentionPolicy struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	
	// Data types to apply this policy to
	ApplyToInteractions bool `json:"apply_to_interactions"`
	ApplyToCases       bool `json:"apply_to_cases"`
	ApplyToPayloads     bool `json:"apply_to_payloads"`
	ApplyToEvidence     bool `json:"apply_to_evidence"`
	ApplyToLogs        bool `json:"apply_to_logs"`
	
	// Retention settings
	RetentionDays int `json:"retention_days"` // How many days to keep data
	MaxRecords    int `json:"max_records"`    // Maximum number of records to keep (0 = unlimited)
	
	// Archive settings
	ArchiveAfterDays   int  `json:"archive_after_days"`   // Days before archiving (0 = never archive)
	ArchiveToStorage   string `json:"archive_to_storage"`   // Archive storage location
	DeleteAfterArchive bool `json:"delete_after_archive"`  // Delete original after archive
	
	// Schedule
	RunHourly   bool `json:"run_hourly"`   // Run hourly
	RunDaily    bool `json:"run_daily"`    // Run daily
	RunWeekly   bool `json:"run_weekly"`   // Run weekly
	RunMonthly  bool `json:"run_monthly"`  // Run monthly
	
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"updated"`
	LastRunAt   *time.Time `json:"last_run_at"`
}

// TableName returns the table name for RetentionPolicy
func (RetentionPolicy) TableName() string {
	return "retention_policies"
}

// Archive represents an archived data set
type Archive struct {
	ID           string    `json:"id" xorm:"'id' pk"`
	PolicyID     string    `json:"policy_id"`
	DataType     string    `json:"data_type"` // interactions, cases, payloads, evidence, logs
	RecordCount  int       `json:"record_count"`
	StoragePath  string    `json:"storage_path"`
	FileSize     int64     `json:"file_size"` // in bytes
	Checksum     string    `json:"checksum"`   // SHA256 checksum
	Compression  string    `json:"compression"` // gzip, zip, none
	
	// Time range
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	
	// Status
	Status       string    `json:"status"` // pending, in_progress, completed, failed
	ErrorMessage string    `json:"error_message"`
	
	CreatedAt    time.Time `json:"created_at" xorm:"created"`
	CompletedAt  *time.Time `json:"completed_at"`
}

// TableName returns the table name for Archive
func (Archive) TableName() string {
	return "archives"
}

// RetentionJob represents a retention job execution
type RetentionJob struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	PolicyID    string    `json:"policy_id"`
	
	// Job details
	JobType     string    `json:"job_type"` // retention, archive, cleanup
	Status      string    `json:"status"`   // pending, running, completed, failed
	
	// Results
	RecordsProcessed int    `json:"records_processed"`
	RecordsDeleted    int    `json:"records_deleted"`
	RecordsArchived   int    `json:"records_archived"`
	
	// Error handling
	ErrorMessage string `json:"error_message"`
	
	// Timing
	StartedAt   time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Duration    int64     `json:"duration"` // in milliseconds
	
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
}

// TableName returns the table name for RetentionJob
func (RetentionJob) TableName() string {
	return "retention_jobs"
}
