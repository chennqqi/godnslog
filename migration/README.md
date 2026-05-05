# Data Migration

This package handles migration from 1.0 data models to 2.0 unified data models.

## Migration Overview

### What is migrated?
- `TblDns` -> `Interaction` (Type=dns)
- `TblHttp` -> `Interaction` (Type=http)

### What is NOT migrated?
- `TblUser` - Kept as-is with wrapper in `internal/models.User`
- `TblResolve` - Kept as-is with wrapper in `internal/models.Resolve`

## Usage

### Basic Migration

```go
import (
    "github.com/chennqqi/godnslog/migration"
    "xorm.io/xorm"
)

func main() {
    engine, _ := xorm.NewEngine("sqlite3", "godnslog.db")
    
    migrator := migration.NewMigrator(engine)
    
    // Perform migration
    err := migrator.MigrateAll()
    if err != nil {
        log.Fatal(err)
    }
    
    // Validate migration
    err = migrator.ValidateMigration()
    if err != nil {
        log.Fatal(err)
    }
}
```

### Rollback

```go
// Rollback all migrations
err := migrator.RollbackAll()
if err != nil {
    log.Fatal(err)
}

// Rollback specific migration
err := migrator.RollbackDNS()
if err != nil {
    log.Fatal(err)
}
```

## Migration Process

### Step 1: Backup
Always backup your database before migration:
```bash
cp godnslog.db godnslog.db.backup
```

### Step 2: Run Migration
```bash
go run cmd/migrate/main.go
```

### Step 3: Validate
The migrator will automatically validate the migration by comparing:
- TblDns count vs DNS Interaction count
- TblHttp count vs HTTP Interaction count

### Step 4: Test
Test the application to ensure all functionality works correctly.

## Rollback

If migration fails or causes issues, you can rollback:
1. Rollback using the migrator
2. Restore from backup

## Notes

- Migration is performed in batches (default 1000 records per batch)
- TblDns and TblHttp records are NOT deleted after migration
- You can safely re-run migration (it will create duplicates)
- Rollback removes all migrated Interaction records

## Future Improvements

- Add option to delete source records after migration
- Add progress bar
- Add dry-run mode
- Add conflict resolution
