package storieshttp

import (
	"github.com/google/uuid"
	"strings"
	"testing"
	"time"
)

func TestImportCommentIDsAreTenantStoryAndSourceScoped(t *testing.T) {
	workspaceID, storyID := uuid.New(), uuid.New()
	first := importCommentID(workspaceID, storyID, "source-1")
	if first != importCommentID(workspaceID, storyID, "source-1") {
		t.Fatal("retry ID changed")
	}
	if first == importCommentID(uuid.New(), storyID, "source-1") || first == importCommentID(workspaceID, uuid.New(), "source-1") || first == importCommentID(workspaceID, storyID, "source-2") {
		t.Fatal("source IDs crossed scope")
	}
}
func TestImportedCommentsEscapeSourceAttributionAndRetainSourceTime(t *testing.T) {
	timestamp := time.Date(2025, 1, 2, 10, 30, 0, 0, time.UTC)
	content := importCommentContent(AppImportComment{AuthorName: `<img onerror="attack">`, Content: "<script>alert(1)</script>\ndecision", CreatedAt: &timestamp})
	if strings.Contains(content, "<script>") || strings.Contains(content, "<img ") {
		t.Fatalf("source text became HTML: %s", content)
	}
	if !strings.Contains(content, "2025-01-02T10:30:00Z") || !strings.Contains(content, "<br>decision") {
		t.Fatalf("source attribution missing: %s", content)
	}
}
func TestImportedRichCommentsDropExecutableHTMLAndPreserveSupportedLinks(t *testing.T) {
	content := importCommentContent(AppImportComment{AuthorName: "Ada", Format: "html", Content: `<p onclick="attack()">Decision<script>attack()</script><a href="javascript:attack()">bad</a><a href="https://example.com/source">source</a><span data-type="mention" data-id="other-user">@someone</span></p>`})
	for _, unsafe := range []string{"onclick", "<script", "javascript:", "data-type", "data-id"} {
		if strings.Contains(content, unsafe) {
			t.Fatalf("unsafe content retained: %s", content)
		}
	}
	if !strings.Contains(content, `href="https://example.com/source"`) || !strings.Contains(content, "Decision") {
		t.Fatalf("supported source content lost: %s", content)
	}
}
