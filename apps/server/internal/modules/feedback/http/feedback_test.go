package feedbackhttp

import (
	"context"
	"errors"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
)

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
	handler := New(nil, resolver, nil)
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
	handler := New(nil, resolver, nil)
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
	handler := New(nil, resolver, nil)
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
