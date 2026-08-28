package taskhandlers

import (
	"context"
	"fmt"
	"testing"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/google/uuid"
)

type mayaAssignmentRead struct {
	actorID       uuid.UUID
	workspaceID   uuid.UUID
	afterStoryID  uuid.UUID
	requestedSize int
}

type fakeMayaAssignmentStore struct {
	assignmentCandidates []mayadomain.AssignmentCandidateStory
	workspacePages       [][]mayadomain.AssignmentCandidateStory
	assignmentReads      []mayaAssignmentRead
	workspaceReads       []mayaAssignmentRead
	assignmentErr        error
	workspaceErr         error
}

func (f *fakeMayaAssignmentStore) ListMayaAssignmentCandidates(
	_ context.Context,
	actorID uuid.UUID,
	afterStoryID uuid.UUID,
	limit int,
) ([]mayadomain.AssignmentCandidateStory, error) {
	f.assignmentReads = append(f.assignmentReads, mayaAssignmentRead{
		actorID:       actorID,
		afterStoryID:  afterStoryID,
		requestedSize: limit,
	})
	return append([]mayadomain.AssignmentCandidateStory(nil), f.assignmentCandidates...), f.assignmentErr
}

func (f *fakeMayaAssignmentStore) ListWorkspaceScheduleCandidates(
	_ context.Context,
	workspaceID uuid.UUID,
	afterStoryID uuid.UUID,
	limit int,
) ([]mayadomain.AssignmentCandidateStory, error) {
	f.workspaceReads = append(f.workspaceReads, mayaAssignmentRead{
		workspaceID:   workspaceID,
		afterStoryID:  afterStoryID,
		requestedSize: limit,
	})
	if f.workspaceErr != nil {
		return nil, f.workspaceErr
	}
	pageIndex := len(f.workspaceReads) - 1
	if pageIndex >= len(f.workspacePages) {
		return nil, nil
	}
	return append([]mayadomain.AssignmentCandidateStory(nil), f.workspacePages[pageIndex]...), nil
}

type fakeMayaTaskProcessor struct {
	reconcileInputs []maya.ReconcileScheduleInput
	batchInputs     []maya.ProcessAssignmentBatchInput
}

func (f *fakeMayaTaskProcessor) ReconcileSchedule(_ context.Context, input maya.ReconcileScheduleInput) error {
	f.reconcileInputs = append(f.reconcileInputs, input)
	return nil
}

func (f *fakeMayaTaskProcessor) ProcessAssignmentBatch(
	_ context.Context,
	input maya.ProcessAssignmentBatchInput,
) (maya.ProcessAssignmentBatchResult, error) {
	input.StoryIDs = append([]uuid.UUID(nil), input.StoryIDs...)
	f.batchInputs = append(f.batchInputs, input)
	return maya.ProcessAssignmentBatchResult{Processed: len(input.StoryIDs)}, nil
}

func (*fakeMayaTaskProcessor) RecoverScheduleOwnerships(context.Context, int) (int, error) {
	return 0, nil
}

func TestListMayaAssignmentCandidatesUsesTypedBoundedStore(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	cursor := uuid.New()
	want := []mayadomain.AssignmentCandidateStory{{ID: uuid.New()}}
	store := &fakeMayaAssignmentStore{assignmentCandidates: want}
	handler := &handlers{mayaAssignments: store, systemUserID: actorID}

	got, err := handler.listMayaAssignmentCandidates(t.Context(), cursor, mayaAssignmentBatchPageSize)
	if err != nil {
		t.Fatalf("list Maya assignment candidates: %v", err)
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Fatalf("unexpected candidates: %#v", got)
	}
	if len(store.assignmentReads) != 1 {
		t.Fatalf("assignment reads = %d, want 1", len(store.assignmentReads))
	}
	read := store.assignmentReads[0]
	if read.actorID != actorID || read.afterStoryID != cursor || read.requestedSize != mayaAssignmentBatchPageSize {
		t.Fatalf("unexpected bounded read: %#v", read)
	}
}

func TestProcessWorkspaceScheduleBatchUsesCursorAndReconcilesEachHumanOwnerOnce(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	systemUserID := uuid.New()
	humanUserID := uuid.New()

	firstPage := make([]mayadomain.AssignmentCandidateStory, mayaWorkspaceSchedulePageSize)
	for index := range firstPage {
		firstPage[index] = mayadomain.AssignmentCandidateStory{
			ID:          orderedMayaTestUUID(index + 1),
			WorkspaceID: workspaceID,
			TeamID:      teamID,
			AssigneeID:  humanUserID,
		}
	}
	firstPage[0].AssigneeID = systemUserID
	firstPage[1].AssigneeID = systemUserID
	secondPage := []mayadomain.AssignmentCandidateStory{
		{
			ID:          orderedMayaTestUUID(mayaWorkspaceSchedulePageSize + 1),
			WorkspaceID: workspaceID,
			TeamID:      teamID,
			AssigneeID:  humanUserID,
		},
		{
			ID:          orderedMayaTestUUID(mayaWorkspaceSchedulePageSize + 2),
			WorkspaceID: workspaceID,
			TeamID:      teamID,
			AssigneeID:  systemUserID,
		},
	}

	store := &fakeMayaAssignmentStore{workspacePages: [][]mayadomain.AssignmentCandidateStory{firstPage, secondPage}}
	processor := &fakeMayaTaskProcessor{}
	handler := &handlers{
		mayaService:     processor,
		mayaAssignments: store,
		systemUserID:    systemUserID,
	}

	if err := handler.processWorkspaceScheduleBatch(t.Context(), workspaceID); err != nil {
		t.Fatalf("process workspace schedule batch: %v", err)
	}

	if len(store.workspaceReads) != 2 {
		t.Fatalf("workspace reads = %d, want 2", len(store.workspaceReads))
	}
	if first := store.workspaceReads[0]; first.workspaceID != workspaceID || first.afterStoryID != uuid.Nil || first.requestedSize != mayaWorkspaceSchedulePageSize {
		t.Fatalf("unexpected first workspace read: %#v", first)
	}
	if second := store.workspaceReads[1]; second.afterStoryID != firstPage[len(firstPage)-1].ID || second.requestedSize != mayaWorkspaceSchedulePageSize {
		t.Fatalf("unexpected second workspace read: %#v", second)
	}
	if len(processor.reconcileInputs) != 1 {
		t.Fatalf("owner reconciliations = %d, want 1", len(processor.reconcileInputs))
	}
	reconcile := processor.reconcileInputs[0]
	if reconcile.WorkspaceID == nil || *reconcile.WorkspaceID != workspaceID || reconcile.UserID == nil || *reconcile.UserID != humanUserID {
		t.Fatalf("unexpected reconciliation input: %#v", reconcile)
	}

	processedStoryCount := 0
	for _, input := range processor.batchInputs {
		if len(input.StoryIDs) > mayaWorkspaceAssignmentBatchSize {
			t.Fatalf("assignment batch size = %d, exceeds %d", len(input.StoryIDs), mayaWorkspaceAssignmentBatchSize)
		}
		processedStoryCount += len(input.StoryIDs)
	}
	if processedStoryCount != 3 {
		t.Fatalf("Maya-assigned stories processed = %d, want 3", processedStoryCount)
	}
}

func TestProcessWorkspaceScheduleBatchBoundsTeamBatches(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	systemUserID := uuid.New()
	page := make([]mayadomain.AssignmentCandidateStory, 60)
	for index := range page {
		page[index] = mayadomain.AssignmentCandidateStory{
			ID:          orderedMayaTestUUID(index + 1),
			WorkspaceID: workspaceID,
			TeamID:      teamID,
			AssigneeID:  systemUserID,
		}
	}

	processor := &fakeMayaTaskProcessor{}
	handler := &handlers{
		mayaService: processor,
		mayaAssignments: &fakeMayaAssignmentStore{
			workspacePages: [][]mayadomain.AssignmentCandidateStory{page},
		},
		systemUserID: systemUserID,
	}

	if err := handler.processWorkspaceScheduleBatch(t.Context(), workspaceID); err != nil {
		t.Fatalf("process workspace schedule batch: %v", err)
	}
	if len(processor.batchInputs) != 3 {
		t.Fatalf("assignment batches = %d, want 3", len(processor.batchInputs))
	}
	wantSizes := []int{25, 25, 10}
	for index, input := range processor.batchInputs {
		if len(input.StoryIDs) != wantSizes[index] {
			t.Fatalf("batch %d size = %d, want %d", index, len(input.StoryIDs), wantSizes[index])
		}
	}
}

func orderedMayaTestUUID(value int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", value))
}
