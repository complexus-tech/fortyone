package stories

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type authorizedBulkDeleteRepo struct {
	Repository
	deleteCalled          bool
	deleteStoryID         uuid.UUID
	deletedIDs            []uuid.UUID
	receivedStoryIDs      []uuid.UUID
	receivedWorkspaceID   uuid.UUID
	receivedAuthorization BulkDeleteAuthorization
}

func (r *authorizedBulkDeleteRepo) Delete(_ context.Context, storyID, workspaceID uuid.UUID, authorization BulkDeleteAuthorization) error {
	r.deleteCalled = true
	r.deleteStoryID = storyID
	r.receivedWorkspaceID = workspaceID
	r.receivedAuthorization = authorization
	return nil
}

func (r *authorizedBulkDeleteRepo) BulkDelete(_ context.Context, storyIDs []uuid.UUID, workspaceID uuid.UUID, authorization BulkDeleteAuthorization) ([]uuid.UUID, error) {
	r.receivedStoryIDs = append([]uuid.UUID(nil), storyIDs...)
	r.receivedWorkspaceID = workspaceID
	r.receivedAuthorization = authorization
	return append([]uuid.UUID(nil), r.deletedIDs...), nil
}

func TestBulkDeletePreservesReceiptAndAuthorization(t *testing.T) {
	requestedIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	deletedIDs := requestedIDs[:2]
	workspaceID := uuid.New()
	authorization := BulkDeleteAuthorization{ActorID: uuid.New(), IsAdmin: false}
	repo := &authorizedBulkDeleteRepo{deletedIDs: deletedIDs}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil, nil)

	got, err := service.BulkDelete(context.Background(), requestedIDs, workspaceID, authorization)
	if err != nil {
		t.Fatalf("bulk delete: %v", err)
	}
	if len(got) != len(deletedIDs) || got[0] != deletedIDs[0] || got[1] != deletedIDs[1] {
		t.Fatalf("deleted receipt = %v, want %v", got, deletedIDs)
	}
	if repo.receivedWorkspaceID != workspaceID || repo.receivedAuthorization != authorization {
		t.Fatalf("repository authorization = workspace %s, auth %#v", repo.receivedWorkspaceID, repo.receivedAuthorization)
	}
	if len(repo.receivedStoryIDs) != len(requestedIDs) {
		t.Fatalf("repository story IDs = %v, want %v", repo.receivedStoryIDs, requestedIDs)
	}
}

func TestDeletePassesAuthoritativeActorAuthorization(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	authorization := BulkDeleteAuthorization{ActorID: uuid.New()}
	repo := &authorizedBulkDeleteRepo{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil, nil)

	if err := service.Delete(context.Background(), storyID, workspaceID, authorization); err != nil {
		t.Fatalf("delete story: %v", err)
	}
	if !repo.deleteCalled || repo.deleteStoryID != storyID || repo.receivedWorkspaceID != workspaceID || repo.receivedAuthorization != authorization {
		t.Fatalf(
			"repository delete = called %t, story %s, workspace %s, auth %#v",
			repo.deleteCalled,
			repo.deleteStoryID,
			repo.receivedWorkspaceID,
			repo.receivedAuthorization,
		)
	}
}
