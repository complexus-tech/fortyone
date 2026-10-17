//go:build integration

package attachmentsrepository

import (
	"strings"
	"testing"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestThumbnailRepositoryScopesClaimsVersionsAndDurableCleanup(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	fixture := newAttachmentFixture(t, ctx, postgres.Pool)
	repo := New(postgres.Pool)
	file := createFixtureAttachment(t, ctx, repo, fixture.workspaceA, fixture.userA, attachmentdomain.OptimizationNotRequested)
	item := attachmentdomain.Thumbnail{ID: uuid.New(), AttachmentID: file.ID, WorkspaceID: file.WorkspaceID,
		SourceVersion: strings.Repeat("a", 64), Provider: "aws", Container: "private", BlobName: "thumbnail-a.jpg"}
	wrong := item
	wrong.WorkspaceID = fixture.workspaceB
	_, err := repo.EnsureThumbnail(ctx, wrong)
	require.ErrorIs(t, err, attachmentdomain.ErrNotFound)
	created, err := repo.EnsureThumbnail(ctx, item)
	require.NoError(t, err)
	require.Equal(t, "pending", created.Status)
	claimed, err := repo.ClaimThumbnail(ctx, item.ID, fixture.workspaceB)
	require.NoError(t, err)
	require.False(t, claimed)
	claimed, err = repo.ClaimThumbnail(ctx, item.ID, file.WorkspaceID)
	require.NoError(t, err)
	require.True(t, claimed)
	claimed, err = repo.ClaimThumbnail(ctx, item.ID, file.WorkspaceID)
	require.NoError(t, err)
	require.False(t, claimed)

	// A concurrent original mutation cannot publish an older derivative.
	_, err = postgres.Pool.Exec(ctx, `UPDATE public.attachments SET updated_at = $2 WHERE attachment_id = $1`, file.ID, file.UpdatedAt.Add(time.Second))
	require.NoError(t, err)
	require.ErrorIs(t, repo.CompleteThumbnail(ctx, item, file.UpdatedAt, 1024), attachmentdomain.ErrStateConflict)
	file, err = repo.GetAttachmentByID(ctx, file.ID, file.WorkspaceID)
	require.NoError(t, err)
	newer := item
	newer.ID = uuid.New()
	newer.SourceVersion = strings.Repeat("b", 64)
	newer.BlobName = "thumbnail-b.jpg"
	_, err = repo.EnsureThumbnail(ctx, newer)
	require.NoError(t, err)
	claimed, err = repo.ClaimThumbnail(ctx, newer.ID, newer.WorkspaceID)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, repo.CompleteThumbnail(ctx, newer, file.UpdatedAt, 1024))
	var staleObjects int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM public.attachment_object_deletion_outbox WHERE attachment_id = $1`, item.ID).Scan(&staleObjects))
	require.Equal(t, 1, staleObjects)

	// Original and derivative retire under independent physical-object IDs.
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO public.attachment_object_deletion_outbox
		(attachment_id, workspace_id, storage_provider, container_name, blob_name) VALUES ($1, $2, 'aws', 'private', 'original.jpg')`, file.ID, file.WorkspaceID)
	require.NoError(t, err)
	require.NoError(t, repo.DeleteAttachment(ctx, file.ID, file.WorkspaceID))
	var count int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM public.attachment_object_deletion_outbox WHERE attachment_id = ANY($1)`, []uuid.UUID{file.ID, newer.ID}).Scan(&count))
	require.Equal(t, 2, count)
}
