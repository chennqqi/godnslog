# Data Retention and Archival

## Overview

Enterprise-grade data retention and archival system for managing data lifecycle, compliance, and storage optimization.

## Features

- **Retention Policies**: Define how long different data types should be kept
- **Automated Cleanup**: Schedule automatic deletion of old data
- **Data Archival**: Archive data before deletion for long-term storage
- **Job Tracking**: Monitor retention and archival jobs
- **Flexible Configuration**: Per-data-type policies with customizable schedules

## Data Models

### RetentionPolicy

Defines retention rules for different data types:

- **ApplyToInteractions**: Apply to interaction data
- **ApplyToCases**: Apply to case data
- **ApplyToPayloads**: Apply to payload data
- **ApplyToEvidence**: Apply to evidence data
- **ApplyToLogs**: Apply to log data
- **RetentionDays**: Number of days to keep data
- **MaxRecords**: Maximum number of records to keep (0 = unlimited)
- **ArchiveAfterDays**: Days before archiving (0 = never archive)
- **Schedule**: Run hourly, daily, weekly, or monthly

### Archive

Represents an archived data set:

- **DataType**: Type of data archived (interactions, cases, payloads, etc.)
- **RecordCount**: Number of records in archive
- **StoragePath**: Path to archive file
- **FileSize**: Size of archive in bytes
- **Checksum**: SHA256 checksum for integrity
- **Compression**: Compression method (gzip, zip, none)
- **TimeRange**: Start and end time of archived data

### RetentionJob

Represents a retention job execution:

- **JobType**: retention, archive, or cleanup
- **Status**: pending, running, completed, failed
- **RecordsProcessed**: Number of records processed
- **RecordsDeleted**: Number of records deleted
- **RecordsArchived**: Number of records archived
- **Duration**: Job execution time in milliseconds

## Usage

### Create Retention Policy

```go
policy := &retention.RetentionPolicy{
    Name:                "Default Policy",
    Description:         "Default 90-day retention",
    ApplyToInteractions: true,
    ApplyToCases:        true,
    ApplyToPayloads:     true,
    RetentionDays:       90,
    MaxRecords:          10000,
    ArchiveAfterDays:   30,
    RunDaily:            true,
    IsEnabled:           true,
}

err := service.CreatePolicy(ctx, policy)
```

### Run Retention Policy

```go
job, err := service.RunPolicy(ctx, "policy-1")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Job %s completed: %d records processed, %d deleted\n",
    job.ID, job.RecordsProcessed, job.RecordsDeleted)
```

### Create Archive

```go
archive, err := service.CreateArchive(ctx, "policy-1", "interactions")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Archive %s created with %d records\n", archive.ID, archive.RecordCount)
```

### List Policies

```go
policies, err := service.ListPolicies(ctx)
for _, policy := range policies {
    fmt.Printf("Policy: %s, Retention: %d days, Enabled: %v\n",
        policy.Name, policy.RetentionDays, policy.IsEnabled)
}
```

### List Archives

```go
archives, err := service.ListArchives(ctx)
for _, archive := range archives {
    fmt.Printf("Archive: %s, Type: %s, Size: %d bytes\n",
        archive.ID, archive.DataType, archive.FileSize)
}
```

## Best Practices

1. **Start Conservative**: Begin with longer retention periods and gradually reduce
2. **Archive Before Delete**: Always archive data before deletion
3. **Monitor Jobs**: Regularly check job execution and status
4. **Test Policies**: Test policies on non-production data first
5. **Document Policies**: Maintain documentation for compliance and auditing

## Security Considerations

1. **Archive Storage**: Secure archive storage with proper access controls
2. **Checksum Validation**: Verify archive integrity before deletion
3. **Audit Logs**: Log all retention and archival operations
4. **Backup Archives**: Maintain multiple copies of critical archives
5. **Compliance**: Ensure policies meet regulatory requirements

## Limitations

- Simplified retention logic (placeholder implementation)
- No actual data export in archival
- No compression support in current implementation
- No deduplication
- No encryption for archived data

## Future Enhancements

- Complete retention logic implementation
- Actual data export and compression
- Archive encryption
- Deduplication support
- Compliance reporting
- Policy templates
- Preview mode for policy changes
