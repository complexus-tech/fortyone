package workspacedomain

import (
	"time"

	"github.com/google/uuid"
)

// InactivityCursor is the stable keyset position used by workspace lifecycle
// jobs. Last access time and workspace ID form the same total order as the
// repository queries, so rows removed or updated during a run cannot cause
// later candidates to be skipped.
type InactivityCursor struct {
	LastAccessedAt time.Time
	WorkspaceID    uuid.UUID
	Valid          bool
}

// InactivityWarningQuery describes one bounded warning-candidate page.
type InactivityWarningQuery struct {
	InactiveBefore time.Time
	Cursor         InactivityCursor
	BatchSize      int
}

// InactivityWarningCandidate contains the workspace-owned data required to
// address one warning email. AdminEmails contains active workspace admins in a
// deterministic membership order.
type InactivityWarningCandidate struct {
	WorkspaceID    uuid.UUID
	Name           string
	Slug           string
	LastAccessedAt time.Time
	AdminEmails    []string
}

// InactivityWarningReceipt records the application-clock instant used after a
// warning email is accepted by the mail transport.
type InactivityWarningReceipt struct {
	WorkspaceID    uuid.UUID
	InactiveBefore time.Time
	WarningSentAt  time.Time
}

// InactivityDeletionBatch describes one atomic, bounded inactive-workspace
// deletion transaction. IntegrationLifecycleLockKey coordinates the aggregate
// deletion with concurrent integration installation changes.
type InactivityDeletionBatch struct {
	InactiveBefore              time.Time
	WarningSentBefore           time.Time
	ProcessedAt                 time.Time
	Cursor                      InactivityCursor
	BatchSize                   int
	IntegrationLifecycleLockKey int64
}

// InactivityDeletionResult reports only committed work. CandidateCount also
// includes workspaces deferred for integration credential migration.
type InactivityDeletionResult struct {
	CandidateCount int
	Deleted        int64
	Blocked        int64
	Cursor         InactivityCursor
}

// DeletedWorkspacePurgeCursor is the stable keyset position for the workspace
// trash-retention policy. DeletedAt and WorkspaceID form the same total order
// as the repository candidate query.
type DeletedWorkspacePurgeCursor struct {
	DeletedAt   time.Time
	WorkspaceID uuid.UUID
	Valid       bool
}

// DeletedWorkspacePurgeBatch describes one atomic, bounded purge transaction
// for workspaces whose soft-deletion retention period has elapsed.
type DeletedWorkspacePurgeBatch struct {
	DeletedBefore               time.Time
	ProcessedAt                 time.Time
	Cursor                      DeletedWorkspacePurgeCursor
	BatchSize                   int
	IntegrationLifecycleLockKey int64
}

// DeletedWorkspacePurgeResult reports only committed work. CandidateCount
// includes candidates deferred until their Slack credentials are encrypted.
type DeletedWorkspacePurgeResult struct {
	CandidateCount int
	Deleted        int64
	Blocked        int64
	Cursor         DeletedWorkspacePurgeCursor
}
