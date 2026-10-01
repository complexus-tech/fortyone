package dataexport

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	exportdomain "github.com/complexus-tech/projects-api/internal/modules/dataexport/domain"
	"github.com/google/uuid"
)

type snapshotRepository struct{ snapshot exportdomain.Snapshot }

func (repository snapshotRepository) Snapshot(context.Context, exportdomain.Scope) (exportdomain.Snapshot, error) {
	return repository.snapshot, nil
}

type auditRecorder struct {
	calls int
	err   error
}

func (audit *auditRecorder) RecordExport(context.Context, uuid.UUID, uuid.UUID, int) error {
	audit.calls++
	return audit.err
}

func TestExportRequiresSuccessfulAuditAndChecksActualBytes(t *testing.T) {
	scope := exportdomain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}
	snapshot := exportdomain.Snapshot{TaskCount: 1, Envelope: exportdomain.Envelope{Format: "fortyone-work-export", Version: 1, Analysis: json.RawMessage(`{}`), CustomFields: json.RawMessage(`[]`), TaskData: json.RawMessage(`[]`)}}
	audit := &auditRecorder{err: errors.New("audit unavailable")}
	if _, err := New(snapshotRepository{snapshot}, audit).Export(t.Context(), scope); !errors.Is(err, audit.err) || audit.calls != 1 {
		t.Fatalf("unaudited export returned: %v, audit calls %d", err, audit.calls)
	}
	oversized, err := json.Marshal(strings.Repeat("x", exportdomain.MaximumBytes))
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Envelope.TaskData = oversized
	audit.calls = 0
	if _, err := New(snapshotRepository{snapshot}, audit).Export(t.Context(), scope); !errors.Is(err, exportdomain.ErrTooLarge) || audit.calls != 0 {
		t.Fatalf("oversized export = %v, audit calls %d", err, audit.calls)
	}
}
