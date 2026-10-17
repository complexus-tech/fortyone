package nativepreviews

import (
	"context"
	"encoding/json"
	"errors"
	domain "github.com/complexus-tech/projects-api/internal/modules/nativepreviews/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"testing"
	"time"
)

type repositoryStub struct {
	calls    []bool
	snapshot domain.Snapshot
	err      error
}

func (r *repositoryStub) Read(_ context.Context, _ auth.Actor, _ uuid.UUID, _ domain.Target, value bool) (domain.Snapshot, error) {
	r.calls = append(r.calls, value)
	return r.snapshot, r.err
}

type cacheStub struct {
	values      map[string]domain.Preview
	gets, sets  int
	unavailable bool
}

func (c *cacheStub) Get(_ context.Context, key string, dest any) error {
	c.gets++
	if c.unavailable {
		return errors.New("unavailable")
	}
	value, ok := c.values[key]
	if !ok {
		return errors.New("miss")
	}
	*dest.(*domain.Preview) = value
	return nil
}
func (c *cacheStub) Set(_ context.Context, key string, value any, _ time.Duration) error {
	c.sets++
	if c.unavailable {
		return errors.New("unavailable")
	}
	c.values[key] = value.(domain.Preview)
	return nil
}
func fixture(t *testing.T) (context.Context, uuid.UUID, *repositoryStub, *cacheStub, *Service) {
	t.Helper()
	workspace := uuid.New()
	ctx, err := auth.BindWorkspace(auth.SetUserID(t.Context(), uuid.New()), workspace)
	if err != nil {
		t.Fatal(err)
	}
	repository := &repositoryStub{snapshot: domain.Snapshot{ID: uuid.New(), Version: "v1", Value: json.RawMessage(`{"title":"Private title"}`)}}
	cache := &cacheStub{values: make(map[string]domain.Preview)}
	return ctx, workspace, repository, cache, New(repository, cache)
}
func TestCacheHitAlwaysReauthorizesWithoutHydratingContent(t *testing.T) {
	ctx, workspace, repository, cache, service := fixture(t)
	target := domain.Target{Kind: "story", ID: repository.snapshot.ID}
	first, err := service.Get(ctx, workspace, target, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Get(ctx, workspace, target, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(repository.calls) != 3 || repository.calls[0] || !repository.calls[1] || repository.calls[2] {
		t.Fatalf("read modes=%v", repository.calls)
	}
	if cache.sets != 1 || !first.RefreshedAt.Equal(second.RefreshedAt) {
		t.Fatal("did not reuse content snapshot")
	}
	repository.err = domain.ErrNotFound
	_, err = service.Get(ctx, workspace, target, false)
	if !errors.Is(err, domain.ErrNotFound) || cache.gets != 2 {
		t.Fatalf("revocation consulted shared cache: error=%v reads=%d", err, cache.gets)
	}
}
func TestNewRevisionAndManualRefreshReplaceSnapshot(t *testing.T) {
	ctx, workspace, repository, cache, service := fixture(t)
	target := domain.Target{Kind: "document", ID: repository.snapshot.ID}
	_, err := service.Get(ctx, workspace, target, false)
	if err != nil {
		t.Fatal(err)
	}
	repository.snapshot.Version = "v2"
	repository.snapshot.Value = json.RawMessage(`{"title":"Renamed"}`)
	next, err := service.Get(ctx, workspace, target, false)
	if err != nil || next.Version != "v2" {
		t.Fatalf("updated=%+v err=%v", next, err)
	}
	reads := cache.gets
	_, err = service.Get(ctx, workspace, target, true)
	if err != nil {
		t.Fatal(err)
	}
	if cache.gets != reads || !repository.calls[len(repository.calls)-1] || cache.sets != 3 {
		t.Fatal("manual refresh did not bypass cache")
	}
}
func TestCacheFailureFallsBackToAuthorizedDatabaseRead(t *testing.T) {
	ctx, workspace, repository, cache, service := fixture(t)
	cache.unavailable = true
	preview, err := service.Get(ctx, workspace, domain.Target{Kind: "objective", ID: repository.snapshot.ID}, false)
	if err != nil || len(preview.Value) == 0 || len(repository.calls) != 2 {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
}
func TestActorAndWorkspaceCheckedBeforeCache(t *testing.T) {
	ctx, workspace, repository, cache, service := fixture(t)
	for _, test := range []struct {
		ctx       context.Context
		workspace uuid.UUID
	}{{context.Background(), workspace}, {ctx, uuid.New()}} {
		_, err := service.Get(test.ctx, test.workspace, domain.Target{Kind: "story"}, false)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("err=%v", err)
		}
	}
	if len(repository.calls) != 0 || cache.gets != 0 {
		t.Fatal("unauthorized request reached data or cache")
	}
}
func TestCacheKeysSeparateTenantAndKind(t *testing.T) {
	snapshot := domain.Snapshot{ID: uuid.New(), Version: "v1"}
	workspace := uuid.New()
	base := cacheKey(workspace, "story", snapshot)
	if base == cacheKey(uuid.New(), "story", snapshot) || base == cacheKey(workspace, "document", snapshot) {
		t.Fatal("cache key collision")
	}
	snapshot.Version = "v2"
	if base == cacheKey(workspace, "story", snapshot) {
		t.Fatal("revision absent from key")
	}
}
