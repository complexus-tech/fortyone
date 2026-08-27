package storiesrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestBulkDeletionAuthorizesAllTargetsInsideTransaction(t *testing.T) {
	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read commands.go: %v", err)
	}
	source := string(data)

	for _, contract := range []struct {
		name  string
		start string
		end   string
	}{
		{name: "soft", start: "func (r *repo) BulkDelete(", end: "// HardBulkDelete"},
		{name: "hard", start: "func (r *repo) HardBulkDelete(", end: "type bulkDeleteTarget struct"},
	} {
		t.Run(contract.name, func(t *testing.T) {
			body := sourceBetweenMarkers(t, source, contract.start, contract.end)
			beginIndex := strings.Index(body, "BeginTxx")
			authorizeIndex := strings.Index(body, "authorizeBulkStoryDeletion")
			mutationIndex := strings.Index(body, "UPDATE stories")
			if contract.name == "hard" {
				mutationIndex = strings.Index(body, "DELETE FROM stories")
			}
			commitIndex := strings.Index(body, "tx.Commit()")
			if beginIndex < 0 || authorizeIndex <= beginIndex || mutationIndex <= authorizeIndex || commitIndex <= mutationIndex {
				t.Fatalf("%s deletion must authorize before mutating and commit one transaction", contract.name)
			}
		})
	}

	authorizationBody := sourceBetweenMarkers(t, source, "func authorizeBulkStoryDeletion(", "func uniqueUUIDs(")
	for _, part := range []string{"workspace_id = :workspace_id", "FOR UPDATE", "len(targets) != len(storyIDs)", "target.ReporterID", "stories.ErrDeleteForbidden"} {
		if !strings.Contains(authorizationBody, part) {
			t.Fatalf("bulk authorization is missing %q", part)
		}
	}
}

func TestOrderUUIDSubsetPreservesRequestOrder(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	third := uuid.New()
	got := orderUUIDSubset([]uuid.UUID{first, second, third}, []uuid.UUID{third, first})
	if len(got) != 2 || got[0] != first || got[1] != third {
		t.Fatalf("ordered subset = %v, want [%s %s]", got, first, third)
	}
}

func TestSoftBulkDeleteReturnsEveryAuthorizedTargetAfterIdempotentRetry(t *testing.T) {
	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read commands.go: %v", err)
	}
	body := sourceBetweenMarkers(t, string(data), "func (r *repo) BulkDelete(", "// HardBulkDelete")

	for _, part := range []string{
		"authorizeBulkStoryDeletion",
		"AND deleted_at IS NULL",
		"return append([]uuid.UUID(nil), targetIDs...), nil",
	} {
		if !strings.Contains(body, part) {
			t.Fatalf("soft deletion retry contract is missing %q", part)
		}
	}
}

func TestSingleDeleteReusesAuthorizedDesiredStateMutation(t *testing.T) {
	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read commands.go: %v", err)
	}
	body := sourceBetweenMarkers(t, string(data), "func (r *repo) Delete(", "func (r *repo) BulkDelete(")

	for _, part := range []string{
		"r.BulkDelete(ctx, []uuid.UUID{id}, workspaceID, authorization)",
		"len(deletedIDs) != 1",
		"deletedIDs[0] != id",
	} {
		if !strings.Contains(body, part) {
			t.Fatalf("single deletion authorization contract is missing %q", part)
		}
	}
}
