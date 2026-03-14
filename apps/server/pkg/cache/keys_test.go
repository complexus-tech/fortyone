package cache

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestObjectiveListCacheKeyIncludesUserID(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userA := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userB := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	keyA := ObjectiveListCacheKey(workspaceID, userA, "")
	keyB := ObjectiveListCacheKey(workspaceID, userB, "")

	if keyA == keyB {
		t.Fatalf("expected different cache keys for different users, got %q", keyA)
	}

	if !strings.Contains(keyA, userA.String()) {
		t.Fatalf("expected cache key %q to include user id %s", keyA, userA)
	}
}

func TestObjectiveListCacheKeyIncludesFilters(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	filters := "team_id:44444444-4444-4444-4444-444444444444;"

	key := ObjectiveListCacheKey(workspaceID, userID, filters)

	if !strings.Contains(key, filters) {
		t.Fatalf("expected cache key %q to include filters %q", key, filters)
	}
}
