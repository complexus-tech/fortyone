package storiesrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoryRetentionQueriesAreBoundedExplicitAndMediaSafe(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/retention.sql")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")
	for _, contract := range []string{
		"story.deleted_at < sqlc.arg(deleted_before)",
		"story.deleted_at > sqlc.arg(after_deleted_at)",
		"story.id > sqlc.arg(after_story_id)",
		"order by story.deleted_at, story.id",
		"limit cast(sqlc.arg(batch_size) as integer)",
		"for update of story skip locked",
		"from public.story_attachments as story_attachment",
		"attachment.workspace_id = story.workspace_id",
		"union select inline_attachment.attachment_id from public.story_inline_attachments as inline_attachment",
		"from public.story_attachments as story_relation",
		"from public.story_inline_attachments as inline_relation",
		"from public.document_attachments as document_relation",
		"insert into public.attachment_object_deletion_outbox",
		"deletion.claim_token = sqlc.arg(claim_token)",
		"deletion.status = 'processing'",
		"deletion.processing_started_at <= sqlc.arg(lease_expired_before)",
		"for update of deletion skip locked",
		"deletion.completed_at < sqlc.arg(completed_before)",
	} {
		require.Contains(t, query, contract)
	}
	for _, forbidden := range []string{
		"now()",
		"current_timestamp",
		"current_date",
		"interval '",
		" offset ",
		"select *",
	} {
		require.NotContains(t, query, forbidden)
	}
}

func TestInteractiveHardDeleteQueryIsBoundedTenantFencedAndOutboxBacked(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/interactive_hard_delete.sql")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")
	for _, contract := range []string{
		"from public.story_attachments as story_attachment",
		"from public.story_inline_attachments as inline_attachment",
		"story.workspace_id = sqlc.arg(workspace_id)",
		"attachment.workspace_id = story.workspace_id",
		"story_attachment.story_id = any(cast(sqlc.arg(story_ids) as uuid[]))",
		"inline_attachment.story_id = any(cast(sqlc.arg(story_ids) as uuid[]))",
		"order by relation.attachment_id",
		"limit cast(sqlc.arg(maximum_attachment_count) as integer)",
	} {
		require.Contains(t, query, contract)
	}
	require.Equal(t, 2, strings.Count(query, "story.workspace_id = sqlc.arg(workspace_id)"))
	require.Equal(t, 2, strings.Count(query, "attachment.workspace_id = story.workspace_id"))
	for _, forbidden := range []string{"now()", "current_timestamp", "current_date", "interval '", " offset ", "select *"} {
		require.NotContains(t, query, forbidden)
	}

	adapter, err := os.ReadFile("interactive_hard_delete.go")
	require.NoError(t, err)
	adapterSource := string(adapter)
	require.Contains(t, adapterSource, "DeleteUnreferencedStoryRetentionAttachments")
	require.Contains(t, adapterSource, "InsertAttachmentObjectDeletionOutbox")
	require.NotContains(t, strings.ToLower(adapterSource), "connectionstring")
	require.NotContains(t, strings.ToLower(adapterSource), "accesskey")
	require.NotContains(t, strings.ToLower(adapterSource), "secret")
}
