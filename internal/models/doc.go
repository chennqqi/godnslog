package models

// Unified Data Models
//
// This package provides unified data models for GODNSLOG 2.0, consolidating:
// - 1.0 models from models/table.go (TblUser, TblDns,TblHttp, TblResolve)
// - 2.0 models from models/v2.go (TblCase, TblPayload, TblInteraction, TblAPIKey)
// - 2.0 models from internal/*/model.go (Case, Payload, Interaction, APIKey, etc.)
//
// Design Principles:
// 1. Use internal/* models as the standard (UUID-based, modern types)
// 2. Provide wrappers for 1.0 models for backward compatibility
// 3. Support data migration from 1.0 to 2.0 models
// 4. Consistent naming conventions (camelCase for JSON, snake_case for database)
//
// Model Hierarchy:
// - User: User management (from TblUser)
// - Case: Test task/project management
// - Payload: Trackable payload with token
// - Interaction: Unified DNS/HTTP/SMTP/LDAP/SMB/FTP interactions
// - APIKey: API authentication
// - Resolve: DNS resolution configuration (from TblResolve)
//
// Migration Path:
// - TblDns -> Interaction (Type=dns)
// - TblHttp -> Interaction (Type=http)
// - TblUser -> User (wrapper)
// - TblResolve -> Resolve (wrapper)
