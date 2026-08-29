package migrations

import (
	"strings"
	"testing"
)

func TestStoryCommentTreeIndexMigrationContract(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000166_story_comment_tree_indexes.up.sql")
	if err != nil {
		t.Fatalf("read story comment index migration: %v", err)
	}
	for _, contract := range []string{
		"idx_story_comments_roots_page",
		"(story_id, created_at DESC, comment_id DESC)",
		"WHERE parent_id IS NULL",
		"idx_story_comments_replies_page",
		"(story_id, parent_id, created_at, comment_id)",
		"WHERE parent_id IS NOT NULL",
	} {
		if !strings.Contains(string(forward), contract) {
			t.Errorf("story comment index migration is missing %q", contract)
		}
	}

	rollback, err := FS.ReadFile("000166_story_comment_tree_indexes.down.sql")
	if err != nil {
		t.Fatalf("read story comment index rollback: %v", err)
	}
	if strings.Count(string(rollback), "DROP INDEX IF EXISTS") != 2 {
		t.Fatalf("story comment index rollback must drop exactly two indexes:\n%s", rollback)
	}
	for _, indexName := range []string{
		"public.idx_story_comments_replies_page",
		"public.idx_story_comments_roots_page",
	} {
		if !strings.Contains(string(rollback), indexName) {
			t.Errorf("story comment index rollback is missing %q", indexName)
		}
	}
}
