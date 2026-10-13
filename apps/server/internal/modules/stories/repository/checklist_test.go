package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type checklistFixture struct {
	checklistQueries
	storyID, workspaceID, teamID, statusID      uuid.UUID
	items                                       []storyreadsql.ListStoryChecklistRow
	created                                     *storyreadsql.CreateStoryMutationParams
	createdCount, associations, events, touched int
	allowed                                     bool
	role                                        string
	fail                                        string
	deletedAt, archivedAt                       *time.Time
}

func (f *checklistFixture) AuthorizeSecondaryStoryTargets(_ context.Context, p storyreadsql.AuthorizeSecondaryStoryTargetsParams) ([]storyreadsql.AuthorizeSecondaryStoryTargetsRow, error) {
	if p.WorkspaceID != f.workspaceID || len(p.StoryIds) != 1 || p.StoryIds[0] != f.storyID {
		return nil, nil
	}
	return []storyreadsql.AuthorizeSecondaryStoryTargetsRow{{ID: f.storyID, TeamID: f.teamID, ActorAuthorized: f.allowed, ActorWorkspaceRole: f.role, DeletedAt: f.deletedAt, ArchivedAt: f.archivedAt}}, nil
}
func (f *checklistFixture) ListStoryChecklist(_ context.Context, p storyreadsql.ListStoryChecklistParams) ([]storyreadsql.ListStoryChecklistRow, error) {
	return append([]storyreadsql.ListStoryChecklistRow(nil), f.items...), nil
}
func (f *checklistFixture) GetStoryMutationSnapshot(context.Context, storyreadsql.GetStoryMutationSnapshotParams) (storyreadsql.GetStoryMutationSnapshotRow, error) {
	return storyreadsql.GetStoryMutationSnapshotRow{ID: f.storyID, TeamID: f.teamID, WorkspaceID: f.workspaceID}, nil
}
func (f *checklistFixture) FindChecklistConversionStatus(context.Context, storyreadsql.FindChecklistConversionStatusParams) (uuid.UUID, error) {
	return f.statusID, nil
}
func (f *checklistFixture) NextStorySequence(context.Context, storyreadsql.NextStorySequenceParams) (int32, error) {
	return 42, nil
}
func (f *checklistFixture) CreateStoryMutation(_ context.Context, p storyreadsql.CreateStoryMutationParams) (storyreadsql.CreateStoryMutationRow, error) {
	if f.fail == "create" {
		return storyreadsql.CreateStoryMutationRow{}, fmt.Errorf("create failed")
	}
	f.created = &p
	f.createdCount++
	return storyreadsql.CreateStoryMutationRow{ID: p.StoryID, SequenceID: p.SequenceID, TeamID: p.TeamID, Title: p.Title, StatusID: p.StatusID}, nil
}
func (f *checklistFixture) InsertStoryAssociation(_ context.Context, p storyreadsql.InsertStoryAssociationParams) (storyreadsql.InsertStoryAssociationRow, error) {
	if f.fail == "association" {
		return storyreadsql.InsertStoryAssociationRow{}, fmt.Errorf("association failed")
	}
	if p.AssociationType != "related" || p.FromStoryID != f.storyID || f.created == nil || p.ToStoryID != f.created.StoryID {
		return storyreadsql.InsertStoryAssociationRow{}, fmt.Errorf("incorrect conversion relationship")
	}
	f.associations++
	return storyreadsql.InsertStoryAssociationRow{}, nil
}
func (f *checklistFixture) CompleteStoryChecklistConversion(_ context.Context, p storyreadsql.CompleteStoryChecklistConversionParams) error {
	if f.fail == "complete" {
		return fmt.Errorf("completion failed")
	}
	for i := range f.items {
		if f.items[i].ID == p.ID {
			f.items[i].Completed = true
			f.items[i].ConvertedAt = p.Now
			f.items[i].ConvertedID = p.ConvertedStoryID
			f.items[i].ConvertedSequenceID = f.created.SequenceID
			f.items[i].ConvertedTitle = &f.created.Title
			f.items[i].ConvertedTeamID = &f.created.TeamID
			f.items[i].ConvertedStatusID = f.created.StatusID
		}
	}
	return nil
}
func (f *checklistFixture) InsertStoryMutationEvent(_ context.Context, p storyreadsql.InsertStoryMutationEventParams) error {
	if f.fail == "event" && p.StoryID == f.storyID {
		return fmt.Errorf("parent event failed")
	}
	f.events++
	return nil
}
func (f *checklistFixture) UpsertStoryMutationActivity(context.Context, storyreadsql.UpsertStoryMutationActivityParams) (storyreadsql.UpsertStoryMutationActivityRow, error) {
	return storyreadsql.UpsertStoryMutationActivityRow{}, nil
}
func (f *checklistFixture) TouchChecklistStory(context.Context, storyreadsql.TouchChecklistStoryParams) error {
	f.touched++
	return nil
}
func (f *checklistFixture) InsertStoryChecklistItem(_ context.Context, p storyreadsql.InsertStoryChecklistItemParams) error {
	f.items = append(f.items, storyreadsql.ListStoryChecklistRow{ID: p.ID, StoryID: p.StoryID, Title: p.Title, Position: p.Position, CreatedAt: p.Now, UpdatedAt: p.Now})
	return nil
}
func (f *checklistFixture) ReorderStoryChecklist(_ context.Context, p storyreadsql.ReorderStoryChecklistParams) (int64, error) {
	ordered := make([]storyreadsql.ListStoryChecklistRow, 0, len(p.ItemIds))
	for i, id := range p.ItemIds {
		for _, row := range f.items {
			if row.ID == id {
				row.Position = int32(i)
				ordered = append(ordered, row)
			}
		}
	}
	f.items = ordered
	return int64(len(ordered)), nil
}
func (f *checklistFixture) ClearStoryChecklist(context.Context, storyreadsql.ClearStoryChecklistParams) error {
	f.items = nil
	return nil
}

func newChecklistFixture(t *testing.T) (*repo, *checklistFixture, storydomain.ChecklistCommand) {
	t.Helper()
	f := &checklistFixture{storyID: uuid.New(), workspaceID: uuid.New(), teamID: uuid.New(), statusID: uuid.New(), allowed: true, role: "member"}
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(f.workspaceID)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	itemID := uuid.New()
	f.items = []storyreadsql.ListStoryChecklistRow{{ID: itemID, StoryID: f.storyID, Title: "Ship it", Position: 0, CreatedAt: now, UpdatedAt: now}}
	command := storydomain.ChecklistCommand{Scope: storydomain.MutationScope{Actor: actor, WorkspaceID: f.workspaceID, ActivityUser: &actor.PrincipalID}, StoryID: f.storyID, ItemID: itemID, Action: storydomain.ChecklistConvert, Event: storydomain.MutationEvent{ID: uuid.New(), StoryID: f.storyID, WorkspaceID: f.workspaceID, Type: storydomain.MutationEventStoryUpdated, Actor: actor, Payload: []byte(`{"changes":{"checklist_changed":true}}`), OccurredAt: now}}
	var mutex sync.Mutex
	r := &repo{runChecklistTransaction: func(ctx context.Context, operation func(checklistQueries) error) error {
		// Model the owning story's FOR UPDATE lock and transactional rollback.
		mutex.Lock()
		defer mutex.Unlock()
		working := *f
		working.items = append([]storyreadsql.ListStoryChecklistRow(nil), f.items...)
		if err := operation(&working); err != nil {
			return err
		}
		*f = working
		return nil
	}}
	return r, f, command
}

func TestChecklistConcurrentConversionCreatesOneStandaloneRelatedStory(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	const count = 12
	results := make(chan storydomain.ChecklistMutationResult, count)
	errs := make(chan error, count)
	var group sync.WaitGroup
	for i := 0; i < count; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := r.ApplyChecklistMutation(context.Background(), command)
			results <- result
			errs <- err
		}()
	}
	group.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var createdID uuid.UUID
	changed := 0
	for result := range results {
		if result.Changed {
			changed++
		}
		if len(result.Items) != 1 || !result.Items[0].Completed || result.Items[0].ConvertedStory == nil {
			t.Fatalf("conversion lost item: %+v", result)
		}
		id := result.Items[0].ConvertedStory.ID
		if createdID == uuid.Nil {
			createdID = id
		} else if id != createdID {
			t.Fatal("conversion retry created another story")
		}
	}
	if changed != 1 || f.createdCount != 1 || f.associations != 1 || f.events != 2 {
		t.Fatalf("side effects repeated: changed=%d fixture=%+v", changed, f)
	}
	if f.created.ParentID != nil || f.created.AssigneeID != nil || f.created.AutoSchedulingStatus != storydomain.AutoSchedulingStatusOff || f.created.StatusID == nil || *f.created.StatusID != f.statusID {
		t.Fatalf("conversion inherited incorrect story state: %+v", f.created)
	}
}

func TestChecklistConversionRollsBackAllEffectsOnFailure(t *testing.T) {
	for _, stage := range []string{"create", "association", "complete", "event"} {
		t.Run(stage, func(t *testing.T) {
			r, f, command := newChecklistFixture(t)
			f.fail = stage
			if _, err := r.ApplyChecklistMutation(context.Background(), command); err == nil {
				t.Fatal("expected injected failure")
			}
			if f.createdCount != 0 || f.associations != 0 || f.events != 0 || f.touched != 0 || f.items[0].Completed || f.items[0].ConvertedAt != nil {
				t.Fatalf("partial conversion survived rollback: %+v", f)
			}
		})
	}
}

func TestChecklistMutationRejectsInvisibleGuestArchivedAndForeignTargets(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*checklistFixture, *storydomain.ChecklistCommand)
		want   error
	}{
		{"not authorized", func(f *checklistFixture, c *storydomain.ChecklistCommand) { f.allowed = false }, storydomain.ErrMutationForbidden},
		{"guest", func(f *checklistFixture, c *storydomain.ChecklistCommand) { f.role = "guest" }, storydomain.ErrMutationForbidden},
		{"foreign workspace", func(f *checklistFixture, c *storydomain.ChecklistCommand) { f.workspaceID = uuid.New() }, storydomain.ErrNotFound},
		{"foreign story", func(f *checklistFixture, c *storydomain.ChecklistCommand) { f.storyID = uuid.New() }, storydomain.ErrNotFound},
		{"foreign item", func(f *checklistFixture, c *storydomain.ChecklistCommand) { c.ItemID = uuid.New() }, storydomain.ErrNotFound},
		{"deleted", func(f *checklistFixture, c *storydomain.ChecklistCommand) { now := time.Now(); f.deletedAt = &now }, storydomain.ErrNotFound},
		{"archived", func(f *checklistFixture, c *storydomain.ChecklistCommand) { now := time.Now(); f.archivedAt = &now }, storydomain.ErrInvalidMutation},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, f, command := newChecklistFixture(t)
			test.change(f, &command)
			_, err := r.ApplyChecklistMutation(context.Background(), command)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if f.createdCount != 0 || f.events != 0 {
				t.Fatal("unauthorized side effects")
			}
		})
	}
}

func TestChecklistStaleOrderDoesNotLoseConcurrentAdditions(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	command.Action = storydomain.ChecklistReorder
	command.ItemIDs = []uuid.UUID{command.ItemID}
	f.items = append(f.items, storyreadsql.ListStoryChecklistRow{ID: uuid.New(), StoryID: f.storyID, Title: "Added elsewhere", Position: 1})
	if _, err := r.ApplyChecklistMutation(context.Background(), command); !errors.Is(err, storydomain.ErrMutationConflict) {
		t.Fatalf("got %v", err)
	}
	if len(f.items) != 2 || f.events != 0 {
		t.Fatal("stale ordering changed checklist")
	}
	command.ItemIDs = []uuid.UUID{f.items[1].ID, f.items[0].ID}
	result, err := r.ApplyChecklistMutation(context.Background(), command)
	if err != nil || !result.Changed || result.Items[0].ID != command.ItemIDs[0] || result.Items[0].Position != 0 {
		t.Fatalf("reorder failed: %+v %v", result, err)
	}
}

func TestChecklistDeletedConvertedStoryCannotBeConvertedAgain(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	now := time.Now().UTC()
	f.items[0].ConvertedAt = &now
	f.items[0].Completed = true
	result, err := r.ApplyChecklistMutation(context.Background(), command)
	if err != nil || result.Changed || f.createdCount != 0 {
		t.Fatalf("deleted conversion created again: %+v %v", result, err)
	}
	incomplete := false
	command.Action = storydomain.ChecklistUpdate
	command.Completed = &incomplete
	if _, err := r.ApplyChecklistMutation(context.Background(), command); !errors.Is(err, storydomain.ErrInvalidMutation) {
		t.Fatalf("converted item reopened: %v", err)
	}
}

func TestChecklistAddLimitAndEmptyClear(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	command.Action = storydomain.ChecklistAdd
	title := "One more"
	command.Title = &title
	command.ItemID = uuid.New()
	for len(f.items) < storydomain.MaxChecklistItems {
		f.items = append(f.items, storyreadsql.ListStoryChecklistRow{ID: uuid.New(), StoryID: f.storyID, Title: "Existing", Position: int32(len(f.items))})
	}
	if _, err := r.ApplyChecklistMutation(context.Background(), command); !errors.Is(err, storydomain.ErrInvalidMutation) {
		t.Fatalf("limit not enforced: %v", err)
	}
	command.Action = storydomain.ChecklistClear
	result, err := r.ApplyChecklistMutation(context.Background(), command)
	if err != nil || len(result.Items) != 0 || !result.Changed {
		t.Fatalf("clear failed: %+v %v", result, err)
	}
	result, err = r.ApplyChecklistMutation(context.Background(), command)
	if err != nil || result.Changed || f.events != 1 {
		t.Fatalf("empty clear emitted another event: %+v %v", result, err)
	}
}
