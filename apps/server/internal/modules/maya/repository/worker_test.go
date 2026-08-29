package mayarepository

import (
	"context"
	"strings"
	"testing"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
	"github.com/google/uuid"
)

func TestListMayaWorkFocusEvidenceGroupsLabelsByStory(t *testing.T) {
	t.Parallel()

	firstStoryID := uuid.New()
	secondStoryID := uuid.New()
	description := "Build typed persistence"
	backend := "backend"
	database := "database"
	queries := &workerQueryStub{evidenceRows: []mayasql.ListMayaWorkFocusEvidenceRow{
		{StoryID: firstStoryID, Title: "Migrate repository", Description: &description, Label: &backend},
		{StoryID: firstStoryID, Title: "Migrate repository", Description: &description, Label: &database},
		{StoryID: secondStoryID, Title: "Document API"},
	}}

	evidence, err := newWithQueries(queries).ListMayaWorkFocusEvidence(
		context.Background(),
		mayadomain.WorkFocusMember{WorkspaceID: uuid.New(), TeamID: uuid.New(), UserID: uuid.New()},
		time.Now().Add(-24*time.Hour),
		30,
	)
	if err != nil {
		t.Fatalf("ListMayaWorkFocusEvidence() error = %v", err)
	}
	if len(evidence) != 2 {
		t.Fatalf("ListMayaWorkFocusEvidence() stories = %d, want 2", len(evidence))
	}
	if got := strings.Join(evidence[0].Labels, ","); got != "backend,database" {
		t.Fatalf("first evidence labels = %q, want backend,database", got)
	}
	if evidence[1].Description != "" || len(evidence[1].Labels) != 0 {
		t.Fatalf("second evidence = %#v, want empty optional fields", evidence[1])
	}
}

func TestListMayaAssignmentCandidatesRejectsImpossibleNullAssignee(t *testing.T) {
	t.Parallel()

	queries := &workerQueryStub{assignmentRows: []mayasql.ListMayaAssignmentCandidatesRow{{
		ID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New(), AssigneeID: nil,
	}}}
	_, err := newWithQueries(queries).ListMayaAssignmentCandidates(
		context.Background(), uuid.New(), uuid.Nil, 25,
	)
	if err == nil || !strings.Contains(err.Error(), "has no assignee") {
		t.Fatalf("ListMayaAssignmentCandidates() error = %v, want missing-assignee error", err)
	}
}

func TestSaveMayaInferredWorkFocusRejectsInvalidConfidence(t *testing.T) {
	t.Parallel()

	queries := &workerQueryStub{}
	_, err := newWithQueries(queries).SaveMayaInferredWorkFocus(
		context.Background(),
		mayadomain.WorkFocusMember{WorkspaceID: uuid.New(), TeamID: uuid.New(), UserID: uuid.New()},
		mayadomain.WorkFocusInferenceResult{StoryCount: 10, Confidence: 1.01},
	)
	if err == nil || !strings.Contains(err.Error(), "between zero and one") {
		t.Fatalf("SaveMayaInferredWorkFocus() error = %v, want confidence validation", err)
	}
	if queries.saveCalls != 0 {
		t.Fatalf("SaveMayaInferredWorkFocus() query calls = %d, want 0", queries.saveCalls)
	}
}

type workerQueryStub struct {
	mayasql.Querier
	evidenceRows   []mayasql.ListMayaWorkFocusEvidenceRow
	assignmentRows []mayasql.ListMayaAssignmentCandidatesRow
	saveCalls      int
}

func (q *workerQueryStub) ListMayaWorkFocusEvidence(
	_ context.Context,
	_ mayasql.ListMayaWorkFocusEvidenceParams,
) ([]mayasql.ListMayaWorkFocusEvidenceRow, error) {
	return q.evidenceRows, nil
}

func (q *workerQueryStub) ListMayaAssignmentCandidates(
	_ context.Context,
	_ mayasql.ListMayaAssignmentCandidatesParams,
) ([]mayasql.ListMayaAssignmentCandidatesRow, error) {
	return q.assignmentRows, nil
}

func (q *workerQueryStub) SaveMayaInferredWorkFocus(
	_ context.Context,
	_ mayasql.SaveMayaInferredWorkFocusParams,
) (int64, error) {
	q.saveCalls++
	return 1, nil
}
