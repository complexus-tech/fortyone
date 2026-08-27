package stories

import (
	"os"
	"strings"
	"testing"
)

func TestTerminalStoryLifecycleEnqueuesImmediateScheduleReconciliation(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("stories.go")
	if err != nil {
		t.Fatalf("read stories.go: %v", err)
	}
	source := string(data)

	for _, contract := range []struct {
		name     string
		start    string
		end      string
		repoCall string
		enqueue  string
	}{
		{
			name: "single delete", start: "func (s *Service) Delete(", end: "// Update updates a story.",
			repoCall: "s.repo.Delete(ctx, id, workspaceId, authorization)", enqueue: "s.enqueueStoryScheduleReconcile(ctx, id, workspaceId)",
		},
		{
			name: "bulk delete", start: "func (s *Service) BulkDelete(", end: "// HardBulkDelete",
			repoCall: "s.repo.BulkDelete(ctx, ids, workspaceId, authorization)", enqueue: "s.enqueueStoryScheduleReconcile(ctx, storyID, workspaceId)",
		},
		{
			name: "bulk archive", start: "func (s *Service) BulkArchive(", end: "// BulkUnarchive",
			repoCall: "s.repo.BulkArchive(ctx, ids, workspaceId)", enqueue: "s.enqueueStoryScheduleReconcile(ctx, storyID, workspaceId)",
		},
	} {
		t.Run(contract.name, func(t *testing.T) {
			body := sourceBetween(t, source, contract.start, contract.end)
			repoIndex := strings.Index(body, contract.repoCall)
			enqueueIndex := strings.Index(body, contract.enqueue)
			if repoIndex < 0 || enqueueIndex < 0 || enqueueIndex <= repoIndex {
				t.Fatalf("successful lifecycle mutation must enqueue reconciliation after persistence; body:\n%s", body)
			}
		})
	}
}

func sourceBetween(t *testing.T, source, start, end string) string {
	t.Helper()
	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing %q", start)
	}
	endOffset := strings.Index(source[startIndex+len(start):], end)
	if endOffset < 0 {
		t.Fatalf("source after %q is missing %q", start, end)
	}
	return source[startIndex : startIndex+len(start)+endOffset]
}
