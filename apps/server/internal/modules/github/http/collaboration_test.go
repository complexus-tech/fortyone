package githubhttp

import (
	"errors"
	"net/http"
	"testing"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	"github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/google/uuid"
)

func TestGitHubCommentIdempotencyIDIsScopedAndRejectsAmbiguity(t *testing.T) {
	t.Parallel()
	workspaceID, actorID, resourceID := uuid.New(), uuid.New(), uuid.New()
	request, err := http.NewRequest(http.MethodPost, "/github-comments", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Idempotency-Key", "github-comment-request-0001")

	first, err := githubCommentIdempotencyID(request, workspaceID, actorID, resourceID, "stories.github-comment.create")
	if err != nil || first == nil {
		t.Fatalf("githubCommentIdempotencyID() = (%v, %v)", first, err)
	}
	second, err := githubCommentIdempotencyID(request, workspaceID, actorID, resourceID, "stories.github-comment.create")
	if err != nil || second == nil || *second != *first {
		t.Fatalf("same scope produced (%v, %v), want %v", second, err, first)
	}
	different, err := githubCommentIdempotencyID(request, workspaceID, actorID, uuid.New(), "stories.github-comment.create")
	if err != nil || different == nil || *different == *first {
		t.Fatalf("different resource produced (%v, %v)", different, err)
	}

	request.Header["Idempotency-Key"] = []string{"github-comment-request-0001", "github-comment-request-0002"}
	if _, err := githubCommentIdempotencyID(request, workspaceID, actorID, resourceID, "stories.github-comment.create"); !errors.Is(err, idempotency.ErrInvalidKey) {
		t.Fatalf("duplicate key error = %v", err)
	}
}

func TestGitHubCollaborationErrorStatusDoesNotExposeInfrastructureAsBadRequest(t *testing.T) {
	t.Parallel()
	if status := githubCollaborationErrorStatus(github.ErrGitHubCommentKeyConflict); status != http.StatusConflict {
		t.Fatalf("conflict status = %d", status)
	}
	if status := githubCollaborationErrorStatus(errors.New("database unavailable")); status != http.StatusInternalServerError {
		t.Fatalf("infrastructure status = %d", status)
	}
	if status := githubCollaborationErrorStatus(github.ErrGitHubOAuthExchangeUnavailable); status != http.StatusServiceUnavailable {
		t.Fatalf("OAuth outage status = %d", status)
	}
}
