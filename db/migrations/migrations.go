package migrations

import "embed"

// Files contains versioned SQLite migrations for Tasq databases.
//
//go:embed issue-tracker/*.sql orchestrator/*.sql
var Files embed.FS
