//go:build integration

package storiesrepository

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestSecondaryStoryMutationsAreTenantFencedAtomicAndReplaySafe(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(
		nil,
		postgres.Pool,
		WithAttachmentObjectStorage("aws", "attachments"),
	)
	baseTime := time.Date(2026, time.August, 28, 13, 0, 0, 0, time.UTC)
	scope := mutationScopeForFixture(t, fixture)

	archiveID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime)
	commands := []storydomain.SecondaryLifecycleCommand{
		secondaryLifecycleCommand(t, scope, []uuid.UUID{archiveID}, storydomain.SecondaryMutationArchive, baseTime.Add(time.Minute)),
		secondaryLifecycleCommand(t, scope, []uuid.UUID{archiveID}, storydomain.SecondaryMutationArchive, baseTime.Add(2*time.Minute)),
	}
	start := make(chan struct{})
	type lifecycleOutcome struct {
		result storydomain.SecondaryLifecycleResult
		err    error
	}
	outcomes := make(chan lifecycleOutcome, len(commands))
	var workers sync.WaitGroup
	for _, command := range commands {
		command := command
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			result, err := repository.ApplySecondaryStoryLifecycle(ctx, command)
			outcomes <- lifecycleOutcome{result: result, err: err}
		}()
	}
	close(start)
	workers.Wait()
	close(outcomes)
	changed := 0
	for outcome := range outcomes {
		if outcome.err != nil {
			t.Fatalf("concurrent archive: %v", outcome.err)
		}
		changed += len(outcome.result.ChangedStoryIDs)
	}
	if changed != 1 {
		t.Fatalf("concurrent archive changed %d stories, want exactly one", changed)
	}
	assertSecondaryStoryState(t, ctx, postgres, archiveID, false, true)
	assertMutationEventCount(t, ctx, postgres, archiveID, "story.updated", 1)

	retry := secondaryLifecycleCommand(
		t, scope, []uuid.UUID{archiveID}, storydomain.SecondaryMutationArchive, baseTime.Add(3*time.Minute),
	)
	retryResult, err := repository.ApplySecondaryStoryLifecycle(ctx, retry)
	if err != nil {
		t.Fatalf("replay archive: %v", err)
	}
	if len(retryResult.ChangedStoryIDs) != 0 {
		t.Fatalf("replayed archive changed stories = %v, want none", retryResult.ChangedStoryIDs)
	}
	assertMutationEventCount(t, ctx, postgres, archiveID, "story.updated", 1)

	unarchive := secondaryLifecycleCommand(
		t, scope, []uuid.UUID{archiveID}, storydomain.SecondaryMutationUnarchive, baseTime.Add(4*time.Minute),
	)
	if _, err := repository.ApplySecondaryStoryLifecycle(ctx, unarchive); err != nil {
		t.Fatalf("unarchive story: %v", err)
	}
	assertSecondaryStoryState(t, ctx, postgres, archiveID, false, false)
	assertMutationEventCount(t, ctx, postgres, archiveID, "story.updated", 2)

	softDeleteID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime.Add(5*time.Minute))
	softDelete := secondaryLifecycleCommand(
		t, scope, []uuid.UUID{softDeleteID}, storydomain.SecondaryMutationSoftDelete, baseTime.Add(6*time.Minute),
	)
	if _, err := repository.ApplySecondaryStoryLifecycle(ctx, softDelete); err != nil {
		t.Fatalf("soft delete story: %v", err)
	}
	assertSecondaryStoryState(t, ctx, postgres, softDeleteID, true, false)
	assertMutationEventCount(t, ctx, postgres, softDeleteID, "story.deleted", 1)
	restore := secondaryLifecycleCommand(
		t, scope, []uuid.UUID{softDeleteID}, storydomain.SecondaryMutationRestore, baseTime.Add(7*time.Minute),
	)
	if _, err := repository.ApplySecondaryStoryLifecycle(ctx, restore); err != nil {
		t.Fatalf("restore story: %v", err)
	}
	assertSecondaryStoryState(t, ctx, postgres, softDeleteID, false, false)
	assertMutationEventCount(t, ctx, postgres, softDeleteID, "story.updated", 1)

	hardDeleteID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime.Add(8*time.Minute))
	hardDelete := secondaryLifecycleCommand(
		t, scope, []uuid.UUID{hardDeleteID}, storydomain.SecondaryMutationHardDelete, baseTime.Add(9*time.Minute),
	)
	hardDeleteResult, err := repository.ApplySecondaryStoryLifecycle(ctx, hardDelete)
	if err != nil {
		t.Fatalf("hard delete story: %v", err)
	}
	if !slices.Equal(hardDeleteResult.ChangedStoryIDs, []uuid.UUID{hardDeleteID}) {
		t.Fatalf("hard delete changed stories = %v", hardDeleteResult.ChangedStoryIDs)
	}
	assertMutationRowCount(t, ctx, postgres, "stories", "id", hardDeleteID, 0)
	assertMutationEventCount(t, ctx, postgres, hardDeleteID, "story.deleted", 1)

	foreignFixture := foreignSecondaryMutationFixture(fixture)
	foreignID := createSecondaryMutationStory(t, ctx, repository, foreignFixture, baseTime.Add(10*time.Minute))
	primaryID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime.Add(11*time.Minute))
	crossTenant := secondaryLifecycleCommand(
		t, scope, []uuid.UUID{primaryID, foreignID}, storydomain.SecondaryMutationArchive, baseTime.Add(12*time.Minute),
	)
	if _, err := repository.ApplySecondaryStoryLifecycle(ctx, crossTenant); !errors.Is(err, storydomain.ErrNotFound) {
		t.Fatalf("cross-tenant batch error = %v, want not found", err)
	}
	assertSecondaryStoryState(t, ctx, postgres, primaryID, false, false)
	assertMutationEventCount(t, ctx, postgres, primaryID, "story.updated", 0)

	conflictID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime.Add(13*time.Minute))
	conflict := secondaryLifecycleCommand(
		t, scope, []uuid.UUID{conflictID}, storydomain.SecondaryMutationArchive, baseTime.Add(14*time.Minute),
	)
	conflict.Events[0].ID = storyCreateMutationCommand(t, fixture, conflictID, baseTime).Event.ID
	// The event identity must collide with a durable row, while remaining valid
	// for this command. Reuse the create event already persisted for the story.
	if err := postgres.Pool.QueryRow(
		ctx,
		"SELECT event_id FROM story_mutation_events WHERE story_id = $1 AND event_type = 'story.created'",
		conflictID,
	).Scan(&conflict.Events[0].ID); err != nil {
		t.Fatalf("load existing mutation event identity: %v", err)
	}
	if _, err := repository.ApplySecondaryStoryLifecycle(ctx, conflict); !errors.Is(err, storydomain.ErrMutationConflict) {
		t.Fatalf("event conflict error = %v, want mutation conflict", err)
	}
	assertSecondaryStoryState(t, ctx, postgres, conflictID, false, false)
	assertMutationEventCount(t, ctx, postgres, conflictID, "story.updated", 0)

	revokedID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime.Add(15*time.Minute))
	mustMutationExec(
		t, ctx, postgres.Pool,
		"DELETE FROM team_members WHERE team_id = $1 AND user_id = $2",
		fixture.teamID, fixture.actorID,
	)
	revoked := secondaryLifecycleCommand(
		t, scope, []uuid.UUID{revokedID}, storydomain.SecondaryMutationArchive, baseTime.Add(16*time.Minute),
	)
	if _, err := repository.ApplySecondaryStoryLifecycle(ctx, revoked); !errors.Is(err, storydomain.ErrMutationForbidden) {
		t.Fatalf("revoked membership error = %v, want forbidden", err)
	}
	assertSecondaryStoryState(t, ctx, postgres, revokedID, false, false)
	mustMutationExec(
		t, ctx, postgres.Pool,
		"INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)",
		fixture.teamID, fixture.actorID,
	)

	machineDeleteID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime.Add(17*time.Minute))
	machineScope, _ := seedStoryMutationServiceAccount(t, ctx, postgres.Pool, fixture, baseTime.Add(-time.Hour))
	machineDelete := secondaryLifecycleCommand(
		t, machineScope, []uuid.UUID{machineDeleteID}, storydomain.SecondaryMutationHardDelete, baseTime.Add(18*time.Minute),
	)
	if _, err := repository.ApplySecondaryStoryLifecycle(ctx, machineDelete); !errors.Is(err, storydomain.ErrMutationForbidden) {
		t.Fatalf("service-account hard-delete error = %v, want forbidden", err)
	}
	assertMutationRowCount(t, ctx, postgres, "stories", "id", machineDeleteID, 1)
}

func TestSecondaryStoryRelationshipsAreScopedAtomicAndEventBacked(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	baseTime := time.Date(2026, time.August, 28, 14, 0, 0, 0, time.UTC)
	scope := mutationScopeForFixture(t, fixture)
	storyID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime)

	labelID, replacementLabelID, foreignLabelID := uuid.New(), uuid.New(), uuid.New()
	mustMutationExec(
		t, ctx, postgres.Pool,
		"INSERT INTO labels (label_id, name, team_id, workspace_id) VALUES ($1, 'Primary', $2, $3), ($4, 'Replacement', $2, $3), ($5, 'Foreign', $6, $7)",
		labelID, fixture.teamID, fixture.workspaceID,
		replacementLabelID, foreignLabelID, fixture.foreignTeamID, fixture.foreignWorkspaceID,
	)
	labelEvent := secondaryMutationEvent(t, scope, storyID, storydomain.MutationEventStoryUpdated, baseTime.Add(time.Minute))
	labelResult, err := repository.ReplaceStoryLabels(ctx, storydomain.ReplaceStoryLabelsCommand{
		Scope: scope, StoryID: storyID, LabelIDs: []uuid.UUID{labelID}, Event: labelEvent,
		Activity: secondaryReplacementActivity(t, scope, storyID, "labels", []uuid.UUID{labelID}, labelEvent.OccurredAt),
	})
	if err != nil || !labelResult.Changed {
		t.Fatalf("replace story labels: result=%#v error=%v", labelResult, err)
	}
	assertSecondaryRelationshipIDs(t, ctx, postgres, "labels", storyID, []uuid.UUID{labelID})
	assertMutationEventCount(t, ctx, postgres, storyID, "story.updated", 1)
	assertSecondaryActivityCount(t, ctx, postgres, storyID, "labels", 1)

	invalidLabelEvent := secondaryMutationEvent(t, scope, storyID, storydomain.MutationEventStoryUpdated, baseTime.Add(2*time.Minute))
	_, err = repository.ReplaceStoryLabels(ctx, storydomain.ReplaceStoryLabelsCommand{
		Scope: scope, StoryID: storyID,
		LabelIDs: []uuid.UUID{replacementLabelID, foreignLabelID}, Event: invalidLabelEvent,
		Activity: secondaryReplacementActivity(
			t, scope, storyID, "labels", []uuid.UUID{replacementLabelID, foreignLabelID}, invalidLabelEvent.OccurredAt,
		),
	})
	if !errors.Is(err, storydomain.ErrInvalidMutation) {
		t.Fatalf("cross-tenant label replacement error = %v, want invalid mutation", err)
	}
	assertSecondaryRelationshipIDs(t, ctx, postgres, "labels", storyID, []uuid.UUID{labelID})
	assertMutationEventCount(t, ctx, postgres, storyID, "story.updated", 1)
	assertSecondaryActivityCount(t, ctx, postgres, storyID, "labels", 1)

	noOpLabelEvent := secondaryMutationEvent(t, scope, storyID, storydomain.MutationEventStoryUpdated, baseTime.Add(3*time.Minute))
	noOpLabels, err := repository.ReplaceStoryLabels(ctx, storydomain.ReplaceStoryLabelsCommand{
		Scope: scope, StoryID: storyID, LabelIDs: []uuid.UUID{labelID, labelID}, Event: noOpLabelEvent,
		Activity: secondaryReplacementActivity(t, scope, storyID, "labels", []uuid.UUID{labelID}, noOpLabelEvent.OccurredAt),
	})
	if err != nil || noOpLabels.Changed {
		t.Fatalf("idempotent label replacement: result=%#v error=%v", noOpLabels, err)
	}
	assertMutationEventCount(t, ctx, postgres, storyID, "story.updated", 1)
	assertSecondaryActivityCount(t, ctx, postgres, storyID, "labels", 1)

	collaboratorID := uuid.New()
	insertMutationUser(t, ctx, postgres.Pool, collaboratorID, true)
	mustMutationExec(
		t, ctx, postgres.Pool,
		"INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')",
		fixture.workspaceID, collaboratorID,
	)
	mustMutationExec(
		t, ctx, postgres.Pool,
		"INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)",
		fixture.teamID, collaboratorID,
	)
	collaboratorEvent := secondaryMutationEvent(t, scope, storyID, storydomain.MutationEventStoryUpdated, baseTime.Add(4*time.Minute))
	collaboratorResult, err := repository.ReplaceStoryCollaborators(ctx, storydomain.ReplaceStoryCollaboratorsCommand{
		Scope: scope, StoryID: storyID, CollaboratorIDs: []uuid.UUID{collaboratorID}, Event: collaboratorEvent,
		Activity: secondaryReplacementActivity(
			t, scope, storyID, "collaborator_ids", []uuid.UUID{collaboratorID}, collaboratorEvent.OccurredAt,
		),
	})
	if err != nil || !collaboratorResult.Changed {
		t.Fatalf("replace story collaborators: result=%#v error=%v", collaboratorResult, err)
	}
	if collaboratorResult.AssigneeID == nil || *collaboratorResult.AssigneeID != fixture.assigneeID {
		t.Fatalf("collaborator replacement assignee = %v, want %s", collaboratorResult.AssigneeID, fixture.assigneeID)
	}
	assertSecondaryRelationshipIDs(t, ctx, postgres, "collaborators", storyID, []uuid.UUID{collaboratorID})
	assertMutationEventCount(t, ctx, postgres, storyID, "story.updated", 2)
	assertSecondaryActivityCount(t, ctx, postgres, storyID, "collaborator_ids", 1)

	invalidCollaboratorEvent := secondaryMutationEvent(t, scope, storyID, storydomain.MutationEventStoryUpdated, baseTime.Add(5*time.Minute))
	_, err = repository.ReplaceStoryCollaborators(ctx, storydomain.ReplaceStoryCollaboratorsCommand{
		Scope: scope, StoryID: storyID,
		CollaboratorIDs: []uuid.UUID{collaboratorID, fixture.assigneeID}, Event: invalidCollaboratorEvent,
		Activity: secondaryReplacementActivity(
			t, scope, storyID, "collaborator_ids", []uuid.UUID{collaboratorID, fixture.assigneeID}, invalidCollaboratorEvent.OccurredAt,
		),
	})
	if !errors.Is(err, storydomain.ErrInvalidMutation) {
		t.Fatalf("assignee collaborator replacement error = %v, want invalid mutation", err)
	}
	assertSecondaryRelationshipIDs(t, ctx, postgres, "collaborators", storyID, []uuid.UUID{collaboratorID})
	assertMutationEventCount(t, ctx, postgres, storyID, "story.updated", 2)
	assertSecondaryActivityCount(t, ctx, postgres, storyID, "collaborator_ids", 1)

	if err := repository.SetStoryWatching(ctx, scope, storyID, fixture.actorID, true); err != nil {
		t.Fatalf("watch story: %v", err)
	}
	assigneeFixture := fixture
	assigneeFixture.actorID = fixture.assigneeID
	assigneeScope := mutationScopeForFixture(t, assigneeFixture)
	if err := repository.SetStoryWatching(ctx, assigneeScope, storyID, fixture.assigneeID, false); err != nil {
		t.Fatalf("mute automatic assignee notifications: %v", err)
	}
	audience, err := repository.ListStoryNotificationAudience(ctx, storyID, fixture.workspaceID)
	if err != nil {
		t.Fatalf("list notification audience: %v", err)
	}
	wantAudience := []uuid.UUID{fixture.actorID, collaboratorID}
	sortSecondaryTestUUIDs(wantAudience)
	if !slices.Equal(audience, wantAudience) {
		t.Fatalf("notification audience = %v, want actor and collaborator", audience)
	}
	assertMutationEventCount(t, ctx, postgres, storyID, "story.updated", 2)

	machineScope, _ := seedStoryMutationServiceAccount(t, ctx, postgres.Pool, fixture, baseTime.Add(-time.Hour))
	if err := repository.SetStoryWatching(
		ctx, machineScope, storyID, machineScope.Actor.PrincipalID, true,
	); !errors.Is(err, storydomain.ErrMutationForbidden) {
		t.Fatalf("machine watch preference error = %v, want forbidden", err)
	}

	mustMutationExec(
		t, ctx, postgres.Pool,
		"DELETE FROM team_members WHERE team_id = $1 AND user_id = $2",
		fixture.teamID, collaboratorID,
	)
	audience, err = repository.ListStoryNotificationAudience(ctx, storyID, fixture.workspaceID)
	if err != nil || !slices.Equal(audience, []uuid.UUID{fixture.actorID}) {
		t.Fatalf("audience after collaborator team revocation = %v error=%v, want actor only", audience, err)
	}
	mustMutationExec(
		t, ctx, postgres.Pool,
		"DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2",
		fixture.workspaceID, fixture.actorID,
	)
	audience, err = repository.ListStoryNotificationAudience(ctx, storyID, fixture.workspaceID)
	if err != nil || len(audience) != 0 {
		t.Fatalf("audience after watcher workspace revocation = %v error=%v, want empty", audience, err)
	}
}

func createSecondaryMutationStory(
	t *testing.T,
	ctx context.Context,
	repository *repo,
	fixture storyMutationFixture,
	createdAt time.Time,
) uuid.UUID {
	t.Helper()
	storyID := uuid.New()
	if _, err := repository.CreateStoryMutation(ctx, storyCreateMutationCommand(t, fixture, storyID, createdAt)); err != nil {
		t.Fatalf("create secondary mutation story: %v", err)
	}
	return storyID
}

func foreignSecondaryMutationFixture(fixture storyMutationFixture) storyMutationFixture {
	foreign := fixture
	foreign.workspaceID = fixture.foreignWorkspaceID
	foreign.teamID = fixture.foreignTeamID
	foreign.statusID = fixture.foreignStatusID
	foreign.actorID = fixture.foreignActorID
	foreign.assigneeID = fixture.foreignActorID
	return foreign
}

func secondaryLifecycleCommand(
	t *testing.T,
	scope storydomain.MutationScope,
	storyIDs []uuid.UUID,
	action storydomain.SecondaryMutationAction,
	changedAt time.Time,
) storydomain.SecondaryLifecycleCommand {
	t.Helper()
	uniqueIDs, err := storydomain.NormalizeSecondaryMutationIDs(storyIDs)
	if err != nil {
		t.Fatalf("normalize secondary lifecycle ids: %v", err)
	}
	events := make([]storydomain.MutationEvent, 0, len(uniqueIDs))
	for _, storyID := range uniqueIDs {
		events = append(events, secondaryMutationEvent(t, scope, storyID, action.EventType(), changedAt))
	}
	return storydomain.SecondaryLifecycleCommand{
		Scope: scope, Action: action, StoryIDs: storyIDs, ChangedAt: changedAt, Events: events,
	}
}

func secondaryMutationEvent(
	t *testing.T,
	scope storydomain.MutationScope,
	storyID uuid.UUID,
	eventType storydomain.MutationEventType,
	occurredAt time.Time,
) storydomain.MutationEvent {
	t.Helper()
	return storydomain.MutationEvent{
		ID: uuid.New(), WorkspaceID: scope.WorkspaceID, StoryID: storyID,
		Type: eventType, Actor: scope.Actor,
		Payload: mustMutationJSON(t, map[string]any{
			"story_id": storyID, "workspace_id": scope.WorkspaceID, "changed": true,
		}),
		OccurredAt: occurredAt,
	}
}

func secondaryReplacementActivity(
	t *testing.T,
	scope storydomain.MutationScope,
	storyID uuid.UUID,
	field string,
	current []uuid.UUID,
	createdAt time.Time,
) *storydomain.MutationActivity {
	t.Helper()
	if scope.ActivityUser == nil {
		t.Fatal("secondary replacement activity requires a user actor")
	}
	return &storydomain.MutationActivity{
		ID: uuid.New(), StoryID: storyID, UserID: *scope.ActivityUser,
		Type: "update", Field: field, CurrentValue: "changed",
		OldValue: mustMutationJSON(t, nil), NewValue: mustMutationJSON(t, current),
		WorkspaceID: scope.WorkspaceID, CreatedAt: createdAt,
	}
}

func assertSecondaryStoryState(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	storyID uuid.UUID,
	deleted, archived bool,
) {
	t.Helper()
	var deletedAt, archivedAt *time.Time
	if err := postgres.Pool.QueryRow(
		ctx, "SELECT deleted_at, archived_at FROM stories WHERE id = $1", storyID,
	).Scan(&deletedAt, &archivedAt); err != nil {
		t.Fatalf("read secondary story lifecycle state: %v", err)
	}
	if (deletedAt != nil) != deleted || (archivedAt != nil) != archived {
		t.Fatalf("story lifecycle deleted=%v archived=%v, want deleted=%v archived=%v", deletedAt, archivedAt, deleted, archived)
	}
}

func assertSecondaryRelationshipIDs(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	kind string,
	storyID uuid.UUID,
	want []uuid.UUID,
) {
	t.Helper()
	queries := map[string]string{
		"labels":        "SELECT label_id FROM story_labels WHERE story_id = $1 ORDER BY label_id",
		"collaborators": "SELECT user_id FROM story_collaborators WHERE story_id = $1 ORDER BY user_id",
	}
	query, ok := queries[kind]
	if !ok {
		t.Fatalf("unsupported secondary relationship kind %q", kind)
	}
	rows, err := postgres.Pool.Query(ctx, query, storyID)
	if err != nil {
		t.Fatalf("list story %s: %v", kind, err)
	}
	defer rows.Close()
	got := make([]uuid.UUID, 0, len(want))
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan story %s: %v", kind, err)
		}
		got = append(got, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate story %s: %v", kind, err)
	}
	sortSecondaryTestUUIDs(want)
	if !slices.Equal(got, want) {
		t.Fatalf("story %s = %v, want %v", kind, got, want)
	}
}

func sortSecondaryTestUUIDs(values []uuid.UUID) {
	slices.SortFunc(values, func(left, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
}

func assertSecondaryActivityCount(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	storyID uuid.UUID,
	field string,
	want int,
) {
	t.Helper()
	var count int
	if err := postgres.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM story_activities WHERE story_id = $1 AND field_changed = $2",
		storyID,
		field,
	).Scan(&count); err != nil {
		t.Fatalf("count %s story activities: %v", field, err)
	}
	if count != want {
		t.Fatalf("%s story activity count = %d, want %d", field, count, want)
	}
}
