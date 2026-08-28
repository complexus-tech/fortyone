package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	"github.com/google/uuid"
)

type mayaWorkFocusEvidenceRead struct {
	member       mayadomain.WorkFocusMember
	updatedAfter time.Time
	limit        int
}

type mayaWorkFocusSave struct {
	member mayadomain.WorkFocusMember
	result mayadomain.WorkFocusInferenceResult
}

type fakeMayaWorkFocusStore struct {
	candidates        []mayadomain.WorkFocusMember
	candidatesErr     error
	candidateLimits   []int
	evidenceByUser    map[uuid.UUID][]mayadomain.WorkFocusEvidence
	evidenceErrByUser map[uuid.UUID]error
	evidenceReads     []mayaWorkFocusEvidenceRead
	saveErrByUser     map[uuid.UUID]error
	saveUpdated       map[uuid.UUID]bool
	saves             []mayaWorkFocusSave
}

func (f *fakeMayaWorkFocusStore) ListMayaWorkFocusCandidates(
	_ context.Context,
	limit int,
) ([]mayadomain.WorkFocusMember, error) {
	f.candidateLimits = append(f.candidateLimits, limit)
	return append([]mayadomain.WorkFocusMember(nil), f.candidates...), f.candidatesErr
}

func (f *fakeMayaWorkFocusStore) ListMayaWorkFocusEvidence(
	_ context.Context,
	member mayadomain.WorkFocusMember,
	updatedAfter time.Time,
	limit int,
) ([]mayadomain.WorkFocusEvidence, error) {
	f.evidenceReads = append(f.evidenceReads, mayaWorkFocusEvidenceRead{
		member:       member,
		updatedAfter: updatedAfter,
		limit:        limit,
	})
	return append([]mayadomain.WorkFocusEvidence(nil), f.evidenceByUser[member.UserID]...), f.evidenceErrByUser[member.UserID]
}

func (f *fakeMayaWorkFocusStore) SaveMayaInferredWorkFocus(
	_ context.Context,
	member mayadomain.WorkFocusMember,
	result mayadomain.WorkFocusInferenceResult,
) (bool, error) {
	f.saves = append(f.saves, mayaWorkFocusSave{member: member, result: result})
	if err := f.saveErrByUser[member.UserID]; err != nil {
		return false, err
	}
	return f.saveUpdated[member.UserID], nil
}

func TestProcessMayaWorkFocusInferenceUsesExplicitTimeAndBoundedReads(t *testing.T) {
	t.Parallel()

	member := mayadomain.WorkFocusMember{
		WorkspaceID: uuid.New(),
		TeamID:      uuid.New(),
		UserID:      uuid.New(),
	}
	evidence := make([]mayadomain.WorkFocusEvidence, 6)
	for index := range evidence {
		evidence[index] = mayadomain.WorkFocusEvidence{Title: "Backend API database work"}
	}
	store := &fakeMayaWorkFocusStore{
		candidates:        []mayadomain.WorkFocusMember{member},
		evidenceByUser:    map[uuid.UUID][]mayadomain.WorkFocusEvidence{member.UserID: evidence},
		evidenceErrByUser: map[uuid.UUID]error{},
		saveErrByUser:     map[uuid.UUID]error{},
		saveUpdated:       map[uuid.UUID]bool{member.UserID: true},
	}
	asOf := time.Date(2026, time.August, 28, 10, 30, 0, 0, time.FixedZone("test", 2*60*60))

	if err := ProcessMayaWorkFocusInference(t.Context(), store, nil, asOf); err != nil {
		t.Fatalf("process Maya work focus inference: %v", err)
	}
	if len(store.candidateLimits) != 1 || store.candidateLimits[0] != mayaWorkFocusMemberBatchSize {
		t.Fatalf("candidate limits = %v, want [%d]", store.candidateLimits, mayaWorkFocusMemberBatchSize)
	}
	if len(store.evidenceReads) != 1 {
		t.Fatalf("evidence reads = %d, want 1", len(store.evidenceReads))
	}
	read := store.evidenceReads[0]
	wantUpdatedAfter := asOf.UTC().Add(-mayaWorkFocusLookback)
	if !read.updatedAfter.Equal(wantUpdatedAfter) || read.limit != mayaWorkFocusEvidenceLimit {
		t.Fatalf("unexpected evidence bounds: %#v", read)
	}
	if len(store.saves) != 1 {
		t.Fatalf("saves = %d, want 1", len(store.saves))
	}
	if result := store.saves[0].result; !result.ShouldInfer || result.StoryCount != len(evidence) || result.RoleTitle == "" {
		t.Fatalf("unexpected inference result: %#v", result)
	}
}

func TestProcessMayaWorkFocusInferenceKeepsPerMemberFailuresBestEffort(t *testing.T) {
	t.Parallel()

	members := []mayadomain.WorkFocusMember{
		{WorkspaceID: uuid.New(), TeamID: uuid.New(), UserID: uuid.New()},
		{WorkspaceID: uuid.New(), TeamID: uuid.New(), UserID: uuid.New()},
		{WorkspaceID: uuid.New(), TeamID: uuid.New(), UserID: uuid.New()},
	}
	evidence := make([]mayadomain.WorkFocusEvidence, 6)
	for index := range evidence {
		evidence[index] = mayadomain.WorkFocusEvidence{Title: "Backend API work"}
	}
	store := &fakeMayaWorkFocusStore{
		candidates: members,
		evidenceByUser: map[uuid.UUID][]mayadomain.WorkFocusEvidence{
			members[1].UserID: evidence,
			members[2].UserID: evidence,
		},
		evidenceErrByUser: map[uuid.UUID]error{members[0].UserID: errors.New("evidence unavailable")},
		saveErrByUser:     map[uuid.UUID]error{members[1].UserID: errors.New("write unavailable")},
		saveUpdated:       map[uuid.UUID]bool{members[2].UserID: true},
	}

	err := ProcessMayaWorkFocusInference(t.Context(), store, nil, time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("best-effort inference returned an error: %v", err)
	}
	if len(store.evidenceReads) != len(members) {
		t.Fatalf("evidence reads = %d, want %d", len(store.evidenceReads), len(members))
	}
	if len(store.saves) != 2 {
		t.Fatalf("save attempts = %d, want 2", len(store.saves))
	}
}

func TestProcessMayaWorkFocusInferenceRejectsMissingDependencies(t *testing.T) {
	t.Parallel()

	if err := ProcessMayaWorkFocusInference(t.Context(), nil, nil, time.Now()); err == nil {
		t.Fatal("expected a missing store error")
	}
	if err := ProcessMayaWorkFocusInference(t.Context(), &fakeMayaWorkFocusStore{}, nil, time.Time{}); err == nil {
		t.Fatal("expected a missing inference time error")
	}
}
