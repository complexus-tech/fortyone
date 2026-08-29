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

func TestInvalidateObjectiveKeysIncludesKeyResultsAndLists(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	objectiveID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	keys := InvalidateObjectiveKeys(workspaceID, objectiveID)

	wantKeyResults := KeyResultsListCacheKey(workspaceID, objectiveID)
	if !containsCacheKey(keys, wantKeyResults) {
		t.Fatalf("expected invalidation keys to contain %q, got %#v", wantKeyResults, keys)
	}
	wantListPattern := "objectives:list:" + workspaceID.String() + "*"
	if !containsCacheKey(keys, wantListPattern) {
		t.Fatalf("expected invalidation keys to contain %q, got %#v", wantListPattern, keys)
	}
}

func TestSecretCacheKeysDoNotContainBearerValues(t *testing.T) {
	t.Parallel()

	secret := "opaque-secret-that-must-not-enter-redis-keys"
	keys := []string{
		AuthSessionCacheKey(secret),
		AuthGoogleStateCacheKey(secret),
		AuthMicrosoftStateCacheKey(secret),
	}

	for _, key := range keys {
		if strings.Contains(key, secret) {
			t.Fatalf("cache key exposed its bearer value: %q", key)
		}
		if !strings.Contains(key, ":v2:") {
			t.Fatalf("cache key does not identify its digest scheme: %q", key)
		}
	}

	firstSessionKey := AuthSessionCacheKey(secret)
	secondSessionKey := AuthSessionCacheKey(strings.Clone(secret))
	if firstSessionKey != secondSessionKey {
		t.Fatal("session cache key must be deterministic")
	}
	if AuthSessionCacheKey(secret) == AuthGoogleStateCacheKey(secret) {
		t.Fatal("different credential purposes must have separate namespaces")
	}
}

func containsCacheKey(keys []string, target string) bool {
	for _, key := range keys {
		if key == target {
			return true
		}
	}
	return false
}
