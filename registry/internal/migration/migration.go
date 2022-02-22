/*
Package migration provides utilities to support the GitLab.com migration, as
described in https://gitlab.com/gitlab-org/container-registry/-/issues/374.
*/
package migration

import "context"

const (
	// CodePathKey is used to get the migration code path from a context.
	CodePathKey = "migration.path"
	// CodePathHeader is used to set/get the migration code path HTTP response header.
	CodePathHeader = "Gitlab-Migration-Path"

	// UnknownCodePath is used when no code path was provided/found.
	UnknownCodePath CodePathVal = iota
	// OldCodePath is used to identify the old code path.
	OldCodePath
	// NewCodePath is used to identify the new code path.
	NewCodePath
)

// CodePathVal is used to define the possible code path values.
type CodePathVal int

// String implements fmt.Stringer.
func (v CodePathVal) String() string {
	switch v {
	case OldCodePath:
		return "old"
	case NewCodePath:
		return "new"
	default:
		return ""
	}
}

type migrationContext struct {
	context.Context
	eligible bool
	path     CodePathVal
}

// Value implements context.Context.
func (mc migrationContext) Value(key interface{}) interface{} {
	switch key {
	case CodePathKey:
		return mc.path
	default:
		return mc.Context.Value(key)
	}
}

// WithCodePath returns a context with the migration code path info.
func WithCodePath(ctx context.Context, path CodePathVal) context.Context {
	return migrationContext{
		Context: ctx,
		path:    path,
	}
}

// CodePath extracts the migration code path info from a context.
func CodePath(ctx context.Context) CodePathVal {
	if v, ok := ctx.Value(CodePathKey).(CodePathVal); ok {
		return v
	}
	return UnknownCodePath
}

// Status enum for repository migration status.
type Status int

// Ineligible statuses.
const (
	StatusUnknown Status = iota // Always first to catch uninitialized values.
	StatusMigrationDisabled
	StatusError
	StatusOldRepo
	StatusNonRepositoryScopedRequest
)

const eligibilityCutoff = 1000

// Eligible statuses.
const (
	StatusNewRepo = iota + eligibilityCutoff
	StatusOnDatabase
)

func (m Status) String() string {
	msg := map[Status]string{
		StatusUnknown:                    "Unknown",
		StatusMigrationDisabled:          "MigrationDisabled",
		StatusError:                      "Error",
		StatusOldRepo:                    "OldRepo",
		StatusNewRepo:                    "NewRepo",
		StatusNonRepositoryScopedRequest: "NonRepositoryScopedRequest",
		StatusOnDatabase:                 "OnDatabase",
	}

	s, ok := msg[m]
	if !ok {
		return msg[StatusUnknown]
	}

	return s
}

// Description returns a human readable description of the migration status.
func (m Status) Description() string {
	msg := map[Status]string{
		StatusUnknown:                    "unknown migration status",
		StatusMigrationDisabled:          "migration mode is disabled in registry config",
		StatusError:                      "error determining migration status",
		StatusOldRepo:                    "repository is old, serving via old code path",
		StatusNewRepo:                    "repository is new, serving via new code path",
		StatusNonRepositoryScopedRequest: "request is not scoped to single repository",
		StatusOnDatabase:                 "repository uses database metadata, serving via new code path",
	}

	s, ok := msg[m]
	if !ok {
		return msg[StatusUnknown]
	}

	return s
}

// ShouldMigrate determines if a repository should be served via the new code path.
func (m Status) ShouldMigrate() bool {
	return m >= eligibilityCutoff
}

// RepositoryStatus represents the status of a repository during an online migration.
type RepositoryStatus string

const (
	// RepositoryStatusNative is the migration status of a repository that was originally created on the metadata database.
	RepositoryStatusNative RepositoryStatus = "native"

	// RepositoryStatusImportInProgress is the migration status of a repository that is currently undergoing an import.
	RepositoryStatusImportInProgress RepositoryStatus = "import_in_progress"

	// RepositoryStatusImportComplete is the migration status of a repository that has successfully been imported.
	RepositoryStatusImportComplete RepositoryStatus = "import_complete"

	// RepositoryStatusImportFailed is the migration status of a repository that has failed to import.
	RepositoryStatusImportFailed RepositoryStatus = "import_failed"

	// RepositoryStatusPreImportInProgress is the migration status of a repository that is currently undergoing a pre import.
	RepositoryStatusPreImportInProgress RepositoryStatus = "pre_import_in_progress"

	// RepositoryStatusPreImportComplete is the migration status of a repository that has successfully been pre imported.
	RepositoryStatusPreImportComplete RepositoryStatus = "pre_import_complete"

	// RepositoryStatusPreImportFailed  the migration status of a repository that has failed to pre import.
	RepositoryStatusPreImportFailed RepositoryStatus = "pre_import_failed"
)

// OnDatabase returns true if the repository uses the database for metadata.
func (s RepositoryStatus) OnDatabase() bool {
	switch s {
	case RepositoryStatusNative, RepositoryStatusImportComplete:
		return true
	default:
		return false
	}
}
