# Internal Package

This package contains the core internal modules for GODNSLOG 2.0.

## Testing

### Run all tests
```bash
go test ./...
```

### Run specific package tests
```bash
go test ./internal/auth
go test ./internal/case
go test ./internal/payload
```

### Run tests with coverage
```bash
go test ./... -cover
```

### Run tests with coverage report
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Structure

- `*_test.go` - Unit tests for each package
- Tests follow Go testing conventions
- Use table-driven tests for multiple scenarios
- Mock dependencies for isolated testing

## Test Coverage Goals

- Core modules (auth, case, payload): 80%+
- Feature modules (canary, rebinding, listener): 70%+
- Utility modules: 60%+
