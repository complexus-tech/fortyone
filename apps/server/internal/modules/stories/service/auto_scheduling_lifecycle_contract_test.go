package stories

import (
	"os"
	"strings"
	"testing"
)

func TestTerminalStoryLifecycleEnqueuesImmediateScheduleReconciliation(t *testing.T) {
	t.Parallel()

	for _, contract := range []struct {
		name     string
		file     string
		start    string
		end      string
		repoCall string
		enqueue  string
	}{
		{
			name: "single delete", file: "story_delete.go", start: "func (s *Service) Delete(",
			repoCall: "mutationRepo.DeleteStoryMutation", enqueue: "s.enqueueStoryScheduleReconcile(ctx, id, workspaceId)",
		},
		{
			name: "bulk delete", file: "story_secondary_lifecycle.go", start: "func (s *Service) BulkDelete(", end: "// HardBulkDelete",
			repoCall: "s.applySecondaryLifecycle(", enqueue: "s.enqueueStoryScheduleReconcile(ctx, storyID, workspaceID)",
		},
		{
			name: "bulk archive", file: "story_secondary_lifecycle.go", start: "func (s *Service) BulkArchive(", end: "// BulkUnarchive",
			repoCall: "s.applySecondaryLifecycle(", enqueue: "s.enqueueStoryScheduleReconcile(ctx, storyID, workspaceID)",
		},
	} {
		t.Run(contract.name, func(t *testing.T) {
			data, err := os.ReadFile(contract.file)
			if err != nil {
				t.Fatalf("read %s: %v", contract.file, err)
			}
			source := string(data)
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
	if end == "" {
		return source[startIndex:]
	}
	endOffset := strings.Index(source[startIndex+len(start):], end)
	if endOffset < 0 {
		t.Fatalf("source after %q is missing %q", start, end)
	}
	return source[startIndex : startIndex+len(start)+endOffset]
}
