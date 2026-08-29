package stories

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type storyMediaReconciliationRepo struct {
	Repository
	story                 CoreSingleStory
	orphanedAttachmentIDs []uuid.UUID
	receivedReferences    []uuid.UUID
	reconcileCalls        int
}

func (r *storyMediaReconciliationRepo) Get(context.Context, uuid.UUID, uuid.UUID) (CoreSingleStory, error) {
	return r.story, nil
}

func (r *storyMediaReconciliationRepo) UpdateWithMediaReconciliation(
	_ context.Context,
	_, _ uuid.UUID,
	_ map[string]any,
	referencedAttachmentIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	r.reconcileCalls++
	r.receivedReferences = append([]uuid.UUID(nil), referencedAttachmentIDs...)
	return append([]uuid.UUID(nil), r.orphanedAttachmentIDs...), nil
}

func TestAuthoritativeStoryMediaReconciliationRunsForUnchangedHTML(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	attachmentID := uuid.New()
	descriptionHTML := `<img data-attachment-id="` + attachmentID.String() + `">`
	orphanedID := uuid.New()
	repo := &storyMediaReconciliationRepo{
		story: CoreSingleStory{
			ID:              storyID,
			Workspace:       workspaceID,
			DescriptionHTML: &descriptionHTML,
		},
		orphanedAttachmentIDs: []uuid.UUID{orphanedID},
	}
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "test"),
		repo,
		nil,
		nil,
	)
	orphanedMediaIDs := []uuid.UUID{}
	err := service.updateWithOptions(
		context.Background(),
		storyID,
		workspaceID,
		uuid.New(),
		map[string]any{"description_html": descriptionHTML},
		updateOptions{
			reconcileMedia:     true,
			referencedMediaIDs: []uuid.UUID{attachmentID},
			orphanedMediaIDs:   &orphanedMediaIDs,
		},
	)
	if err != nil {
		t.Fatalf("reconcile unchanged description: %v", err)
	}
	if repo.reconcileCalls != 1 || len(repo.receivedReferences) != 1 || repo.receivedReferences[0] != attachmentID {
		t.Fatalf("reconcile calls=%d references=%v", repo.reconcileCalls, repo.receivedReferences)
	}
	if len(orphanedMediaIDs) != 1 || orphanedMediaIDs[0] != orphanedID {
		t.Fatalf("orphaned media=%v, want [%s]", orphanedMediaIDs, orphanedID)
	}
}

func TestStoryMediaReconciliationRequiresDescriptionHTML(t *testing.T) {
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "test"),
		&storyMediaReconciliationRepo{},
		nil,
		nil,
	)
	if _, err := service.UpdateWithMediaReconciliation(
		context.Background(),
		uuid.New(),
		uuid.New(),
		map[string]any{"title": "No media snapshot"},
		nil,
	); err != ErrInvalidStoryMediaReference {
		t.Fatalf("error=%v, want invalid story media reference", err)
	}
}
