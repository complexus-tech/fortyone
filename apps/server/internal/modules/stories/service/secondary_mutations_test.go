package stories

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type secondaryMutationRepositoryStub struct {
	Repository
	story               CoreSingleStory
	lifecycleResult     *storydomain.SecondaryLifecycleResult
	lifecycleErr        error
	lifecycleCommand    *storydomain.SecondaryLifecycleCommand
	labelCommand        *storydomain.ReplaceStoryLabelsCommand
	collaboratorCommand *storydomain.ReplaceStoryCollaboratorsCommand
	getCalls            int
}

func (repository *secondaryMutationRepositoryStub) ApplySecondaryStoryLifecycle(
	_ context.Context,
	command storydomain.SecondaryLifecycleCommand,
) (storydomain.SecondaryLifecycleResult, error) {
	repository.lifecycleCommand = &command
	if repository.lifecycleResult != nil || repository.lifecycleErr != nil {
		if repository.lifecycleResult == nil {
			return storydomain.SecondaryLifecycleResult{}, repository.lifecycleErr
		}
		return *repository.lifecycleResult, repository.lifecycleErr
	}
	return storydomain.SecondaryLifecycleResult{
		StoryIDs: append([]uuid.UUID(nil), command.StoryIDs...), ChangedStoryIDs: append([]uuid.UUID(nil), command.StoryIDs...),
	}, nil
}

func TestHardBulkDeletePropagatesDurableAttachmentDeletionOwnership(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, storyID, attachmentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repository := &secondaryMutationRepositoryStub{
		lifecycleResult: &storydomain.SecondaryLifecycleResult{
			StoryIDs:                         []uuid.UUID{storyID},
			ChangedStoryIDs:                  []uuid.UUID{storyID},
			OrphanedAttachmentIDs:            []uuid.UUID{attachmentID},
			AttachmentObjectDeletionDeferred: true,
		},
	}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repository, nil, nil)
	ctx := storyMutationActorContext(t, workspaceID, actorID)

	result, err := service.HardBulkDelete(
		ctx,
		[]uuid.UUID{storyID},
		workspaceID,
		BulkDeleteAuthorization{ActorID: actorID},
	)

	if err != nil {
		t.Fatalf("HardBulkDelete() error = %v", err)
	}
	if !result.AttachmentObjectDeletionDeferred {
		t.Fatal("HardBulkDelete() did not preserve durable attachment-deletion ownership")
	}
	if len(result.OrphanedAttachmentIDs) != 1 || result.OrphanedAttachmentIDs[0] != attachmentID {
		t.Fatalf("HardBulkDelete() retired attachments = %v, want %s", result.OrphanedAttachmentIDs, attachmentID)
	}
	if repository.lifecycleCommand == nil || repository.lifecycleCommand.Action != storydomain.SecondaryMutationHardDelete {
		t.Fatalf("HardBulkDelete() lifecycle command = %#v", repository.lifecycleCommand)
	}
}

func (repository *secondaryMutationRepositoryStub) ReplaceStoryLabels(
	_ context.Context,
	command storydomain.ReplaceStoryLabelsCommand,
) (storydomain.ReplacementResult, error) {
	repository.labelCommand = &command
	return storydomain.ReplacementResult{Changed: true, CurrentIDs: append([]uuid.UUID(nil), command.LabelIDs...)}, nil
}

func (repository *secondaryMutationRepositoryStub) ReplaceStoryCollaborators(
	_ context.Context,
	command storydomain.ReplaceStoryCollaboratorsCommand,
) (storydomain.ReplacementResult, error) {
	repository.collaboratorCommand = &command
	return storydomain.ReplacementResult{
		Changed: true, CurrentIDs: append([]uuid.UUID(nil), command.CollaboratorIDs...),
		AssigneeID: repository.story.Assignee,
	}, nil
}

func (*secondaryMutationRepositoryStub) SetStoryWatching(
	context.Context,
	storydomain.MutationScope,
	uuid.UUID,
	uuid.UUID,
	bool,
) error {
	return nil
}

func (*secondaryMutationRepositoryStub) ListStoryNotificationAudience(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) ([]uuid.UUID, error) {
	return []uuid.UUID{}, nil
}

func (repository *secondaryMutationRepositoryStub) Get(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (CoreSingleStory, error) {
	repository.getCalls++
	return repository.story, nil
}

func (*secondaryMutationRepositoryStub) RecordActivities(
	_ context.Context,
	activities []CoreActivity,
) ([]CoreActivity, error) {
	return activities, nil
}

func TestBuildSecondaryLifecycleCommandUsesMinimalStablePayloads(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	storyA, storyB := uuid.New(), uuid.New()
	changedAt := time.Date(2026, time.August, 28, 17, 0, 0, 0, time.UTC)
	scope := storyMutationScopeForTest(t, workspaceID, actorID)

	archive, err := buildSecondaryLifecycleCommand(
		scope,
		[]uuid.UUID{storyA, storyB, storyA},
		storydomain.SecondaryMutationArchive,
		changedAt,
	)
	if err != nil {
		t.Fatalf("build archive command: %v", err)
	}
	if len(archive.StoryIDs) != 2 || len(archive.Events) != 2 {
		t.Fatalf("normalized archive command = %#v", archive)
	}
	payload := decodeSecondaryMutationPayload(t, archive.Events[0].Payload)
	changes, ok := payload["changes"].(map[string]any)
	if !ok || changes["archived_at"] != changedAt.Format(time.RFC3339) {
		t.Fatalf("archive event changes = %#v", payload["changes"])
	}
	if len(payload) != 3 {
		t.Fatalf("archive event payload keys = %#v, want identity and changes only", payload)
	}

	deleted, err := buildSecondaryLifecycleCommand(
		scope,
		[]uuid.UUID{storyA},
		storydomain.SecondaryMutationHardDelete,
		changedAt,
	)
	if err != nil {
		t.Fatalf("build hard-delete command: %v", err)
	}
	deletedPayload := decodeSecondaryMutationPayload(t, deleted.Events[0].Payload)
	if len(deletedPayload) != 2 || deletedPayload["storyId"] != storyA.String() ||
		deletedPayload["workspaceId"] != workspaceID.String() {
		t.Fatalf("deleted event payload = %#v", deletedPayload)
	}
}

func TestSecondaryRelationshipEventsDoNotExposeMembershipLists(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, storyID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repository := &secondaryMutationRepositoryStub{story: CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: teamID, Title: "Minimal payload",
	}}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repository, nil, nil)
	ctx := storyMutationActorContext(t, workspaceID, actorID)

	labelIDs := []uuid.UUID{uuid.New(), uuid.New()}
	if err := service.UpdateLabels(ctx, storyID, workspaceID, labelIDs); err != nil {
		t.Fatalf("UpdateLabels() error = %v", err)
	}
	if repository.labelCommand == nil {
		t.Fatal("typed label command was not captured")
	}
	if repository.labelCommand.Activity == nil || repository.labelCommand.Activity.Field != "labels" {
		t.Fatalf("typed label activity = %#v", repository.labelCommand.Activity)
	}
	assertSecondaryChangedMarker(t, repository.labelCommand.Event.Payload, "label_ids")

	collaboratorIDs := []uuid.UUID{uuid.New(), uuid.New()}
	if err := service.UpdateCollaborators(ctx, storyID, workspaceID, collaboratorIDs); err != nil {
		t.Fatalf("UpdateCollaborators() error = %v", err)
	}
	if repository.collaboratorCommand == nil {
		t.Fatal("typed collaborator command was not captured")
	}
	if repository.collaboratorCommand.Activity == nil || repository.collaboratorCommand.Activity.Field != "collaborator_ids" {
		t.Fatalf("typed collaborator activity = %#v", repository.collaboratorCommand.Activity)
	}
	if repository.getCalls != 0 {
		t.Fatalf("typed relationship path performed %d compatibility story reads", repository.getCalls)
	}
	assertSecondaryChangedMarker(t, repository.collaboratorCommand.Event.Payload, "collaborator_ids")
}

func storyMutationScopeForTest(t *testing.T, workspaceID, actorID uuid.UUID) storydomain.MutationScope {
	t.Helper()
	ctx := storyMutationActorContext(t, workspaceID, actorID)
	scope, err := mutationScope(ctx, workspaceID, actorID, platformauth.PrincipalHumanUser)
	if err != nil {
		t.Fatalf("build mutation scope: %v", err)
	}
	return scope
}

func decodeSecondaryMutationPayload(t *testing.T, payload json.RawMessage) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode secondary mutation payload: %v", err)
	}
	return decoded
}

func assertSecondaryChangedMarker(t *testing.T, payload json.RawMessage, field string) {
	t.Helper()
	decoded := decodeSecondaryMutationPayload(t, payload)
	changes, ok := decoded["changes"].(map[string]any)
	if !ok || len(decoded) != 3 || len(changes) != 1 || changes[field] != "changed" {
		t.Fatalf("%s event changes = %#v", field, decoded["changes"])
	}
	encoded := string(payload)
	if len(encoded) > 512 {
		t.Fatalf("relationship event payload unexpectedly large: %d bytes", len(encoded))
	}
}
