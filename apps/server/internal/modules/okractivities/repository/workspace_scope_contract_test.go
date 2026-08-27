package okractivitiesrepository

import (
	"os"
	"strings"
	"testing"
)

func TestOKRActivityQueriesScopeOwningWorkspace(t *testing.T) {
	data, err := os.ReadFile("queries.go")
	if err != nil {
		t.Fatalf("read queries.go: %v", err)
	}
	source := string(data)

	objectiveStart := strings.Index(source, "func (r *repo) GetObjectiveActivities(")
	keyResultStart := strings.Index(source, "func (r *repo) GetKeyResultActivities(")
	if objectiveStart < 0 || keyResultStart <= objectiveStart {
		t.Fatal("activity query functions are missing")
	}
	objectiveBody := source[objectiveStart:keyResultStart]
	for _, part := range []string{"JOIN objectives o", "oa.workspace_id = :workspace_id", "o.workspace_id = :workspace_id"} {
		if !strings.Contains(objectiveBody, part) {
			t.Fatalf("objective activity query is missing %q", part)
		}
	}

	keyResultBody := source[keyResultStart:]
	for _, part := range []string{"JOIN key_results kr", "JOIN objectives o", "oa.workspace_id = :workspace_id", "o.workspace_id = :workspace_id"} {
		if !strings.Contains(keyResultBody, part) {
			t.Fatalf("key-result activity query is missing %q", part)
		}
	}
}
