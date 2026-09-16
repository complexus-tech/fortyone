package mayahttp

import (
	"errors"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestRealtimeDeleteRequiresExactCurrentApprovalAndAuthorization(t *testing.T) {
	h := &Handlers{secretKey: "test-only-secret"}
	actor, workspace, session := uuid.New(), uuid.New(), uuid.New()
	story := stories.CoreSingleStory{ID: uuid.New(), Workspace: workspace, Reporter: &actor, Title: "Prepare launch", TeamCode: "ENG", SequenceID: 42, UpdatedAt: time.Now().UTC()}
	authorization := stories.BulkDeleteAuthorization{ActorID: actor}
	calls := 0
	remove := func() error { calls++; return nil }
	preview, err := h.confirmAndDeleteRealtimeStory(session, workspace, story, AppRealtimeDeleteStoryArguments{}, authorization, remove)
	if err != nil || !preview.RequiresConfirmation || preview.Confirmation.Title != story.Title || calls != 0 {
		t.Fatalf("preview=%+v calls=%d err=%v", preview, calls, err)
	}
	args := AppRealtimeDeleteStoryArguments{Confirmed: true, ConfirmationToken: preview.ConfirmationToken}
	for _, name := range []string{"changed story", "changed title", "changed version", "changed session", "wrong workspace", "noncreator", "missing token"} {
		t.Run(name, func(t *testing.T) {
			candidate, candidateSession, candidateWorkspace, candidateAuth, candidateArgs := story, session, workspace, authorization, args
			switch name {
			case "changed story":
				candidate.ID = uuid.New()
			case "changed title":
				candidate.Title = "Another launch"
			case "changed version":
				candidate.UpdatedAt = candidate.UpdatedAt.Add(time.Second)
			case "changed session":
				candidateSession = uuid.New()
			case "wrong workspace":
				candidateWorkspace = uuid.New()
			case "noncreator":
				candidateAuth.ActorID = uuid.New()
			case "missing token":
				candidateArgs.ConfirmationToken = ""
			}
			result, err := h.confirmAndDeleteRealtimeStory(candidateSession, candidateWorkspace, candidate, candidateArgs, candidateAuth, remove)
			if err != nil || result.Success || calls != 0 {
				t.Fatalf("result=%+v calls=%d err=%v", result, calls, err)
			}
		})
	}
	result, err := h.confirmAndDeleteRealtimeStory(session, workspace, story, args, authorization, remove)
	if err != nil || !result.Success || calls != 1 {
		t.Fatalf("result=%+v calls=%d err=%v", result, calls, err)
	}
}

func TestRealtimeDeleteRetainsTransactionalAuthorizationFailure(t *testing.T) {
	h := &Handlers{secretKey: "test-only-secret"}
	workspace, session := uuid.New(), uuid.New()
	story := stories.CoreSingleStory{ID: uuid.New(), Workspace: workspace, Title: "Task"}
	authorization := stories.BulkDeleteAuthorization{ActorID: uuid.New(), IsAdmin: true}
	preview, err := h.confirmAndDeleteRealtimeStory(session, workspace, story, AppRealtimeDeleteStoryArguments{}, authorization, func() error { t.Fatal("preview deleted"); return nil })
	if err != nil {
		t.Fatal(err)
	}
	result, err := h.confirmAndDeleteRealtimeStory(session, workspace, story, AppRealtimeDeleteStoryArguments{Confirmed: true, ConfirmationToken: preview.ConfirmationToken}, authorization, func() error { return stories.ErrDeleteForbidden })
	if !errors.Is(err, stories.ErrDeleteForbidden) || result.Success {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
