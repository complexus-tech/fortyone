package feedbackhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

func TestHTTPStatusClassifiesFeedbackErrors(t *testing.T) {
	t.Parallel()

	if status := httpStatus(fmt.Errorf("%w: invalid status", feedback.ErrInvalidInput)); status != http.StatusBadRequest {
		t.Fatalf("invalid input status = %d, want %d", status, http.StatusBadRequest)
	}
	if status := httpStatus(feedback.ErrStoryManaged); status != http.StatusConflict {
		t.Fatalf("story-managed status = %d, want %d", status, http.StatusConflict)
	}
	if status := httpStatus(errors.New("database unavailable")); status != http.StatusInternalServerError {
		t.Fatalf("unexpected error status = %d, want %d", status, http.StatusInternalServerError)
	}
}

type teamAccessStub struct {
	err error
}

func (s teamAccessStub) GetByID(_ context.Context, teamID, workspaceID, _ uuid.UUID) (teams.CoreTeam, error) {
	if s.err != nil {
		return teams.CoreTeam{}, s.err
	}
	return teams.CoreTeam{ID: teamID, Workspace: workspaceID}, nil
}

func TestAuthorizeTeamMasksMembershipFailuresAsNotFound(t *testing.T) {
	t.Parallel()

	handler := New(nil, teamAccessStub{err: teams.ErrTeamNotFound}, nil, nil)
	err := handler.authorizeTeam(context.Background(), uuid.New(), uuid.New(), uuid.New())

	if !errors.Is(err, feedback.ErrNotFound) {
		t.Fatalf("authorize team error = %v, want not found", err)
	}
}

func TestAuthorizeTeamPreservesInfrastructureFailures(t *testing.T) {
	t.Parallel()

	infrastructureErr := errors.New("database unavailable")
	handler := New(nil, teamAccessStub{err: infrastructureErr}, nil, nil)
	err := handler.authorizeTeam(context.Background(), uuid.New(), uuid.New(), uuid.New())

	if !errors.Is(err, infrastructureErr) {
		t.Fatalf("authorize team error = %v, want infrastructure failure", err)
	}
}

func TestAppPortalDoesNotExposeRemovedDescription(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(toAppPortal(feedback.CorePortal{ID: uuid.New()}))
	if err != nil {
		t.Fatalf("marshal portal: %v", err)
	}
	if strings.Contains(string(payload), "description") {
		t.Fatalf("portal payload unexpectedly contains description: %s", payload)
	}
}

type profileImageResolverStub struct {
	resolved map[string]string
	errors   map[string]error
	calls    map[string]int
}

func (s *profileImageResolverStub) ResolveProfileImageURL(_ context.Context, avatar string, _ time.Duration) (string, error) {
	if s.calls == nil {
		s.calls = make(map[string]int)
	}
	s.calls[avatar]++
	if err := s.errors[avatar]; err != nil {
		return "", err
	}
	return s.resolved[avatar], nil
}

func stringPointer(value string) *string {
	return &value
}

func TestResolvePortalAvatarsUsesPresignedURLs(t *testing.T) {
	t.Parallel()

	itemID := uuid.New()
	resolver := &profileImageResolverStub{
		resolved: map[string]string{
			"profiles/joseph.webp": "https://signed.example.com/profiles/joseph.webp",
		},
	}
	handler := New(nil, nil, resolver, nil)
	portal := feedback.CorePortalSnapshot{
		Items: []feedback.CoreItem{
			{ID: itemID, AuthorAvatar: stringPointer("profiles/joseph.webp")},
		},
		Comments: []feedback.CoreComment{
			{ItemID: itemID, AuthorAvatar: stringPointer("profiles/joseph.webp")},
		},
	}

	handler.resolvePortalAvatars(context.Background(), &portal)

	want := "https://signed.example.com/profiles/joseph.webp"
	if portal.Items[0].AuthorAvatar == nil || *portal.Items[0].AuthorAvatar != want {
		t.Fatalf("item avatar = %v, want %q", portal.Items[0].AuthorAvatar, want)
	}
	if portal.Comments[0].AuthorAvatar == nil || *portal.Comments[0].AuthorAvatar != want {
		t.Fatalf("comment avatar = %v, want %q", portal.Comments[0].AuthorAvatar, want)
	}
	if got := resolver.calls["profiles/joseph.webp"]; got != 1 {
		t.Fatalf("resolver calls = %d, want 1", got)
	}
}

func TestResolvePortalAvatarsOmitsUnresolvableAvatar(t *testing.T) {
	t.Parallel()

	resolver := &profileImageResolverStub{
		errors: map[string]error{
			"missing.webp": errors.New("presign failed"),
		},
	}
	handler := New(nil, nil, resolver, nil)
	portal := feedback.CorePortalSnapshot{
		Items: []feedback.CoreItem{
			{ID: uuid.New(), AuthorAvatar: stringPointer("missing.webp")},
		},
	}

	handler.resolvePortalAvatars(context.Background(), &portal)

	if portal.Items[0].AuthorAvatar != nil {
		t.Fatalf("item avatar = %q, want nil", *portal.Items[0].AuthorAvatar)
	}
}

func TestResolvePortalAvatarsOnlyResolvesVisibleComments(t *testing.T) {
	t.Parallel()

	resolver := &profileImageResolverStub{
		resolved: map[string]string{
			"visible.webp": "https://signed.example.com/visible.webp",
		},
	}
	handler := New(nil, nil, resolver, nil)
	visibleItemID := uuid.New()
	portal := feedback.CorePortalSnapshot{
		Items: []feedback.CoreItem{{ID: visibleItemID}},
		Comments: []feedback.CoreComment{
			{ItemID: visibleItemID, AuthorAvatar: stringPointer("visible.webp")},
			{ItemID: uuid.New(), AuthorAvatar: stringPointer("not-visible.webp")},
		},
	}

	handler.resolvePortalAvatars(context.Background(), &portal)

	if got := resolver.calls["visible.webp"]; got != 1 {
		t.Fatalf("visible avatar resolver calls = %d, want 1", got)
	}
	if got := resolver.calls["not-visible.webp"]; got != 0 {
		t.Fatalf("hidden avatar resolver calls = %d, want 0", got)
	}
}
