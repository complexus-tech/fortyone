//go:build integration

package attachmentsrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryScopesAttachmentLifecycleToWorkspace(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	fixture := newAttachmentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	attachment := createFixtureAttachment(t, ctx, repository, fixture.workspaceA, fixture.userA, attachmentdomain.OptimizationNotRequested)

	if _, err := repository.GetAttachmentByID(ctx, attachment.ID, fixture.workspaceA); err != nil {
		t.Fatalf("get own attachment: %v", err)
	}
	if _, err := repository.GetAttachmentByID(ctx, attachment.ID, fixture.workspaceB); !errors.Is(err, attachmentdomain.ErrNotFound) {
		t.Fatalf("cross-workspace get error = %v", err)
	}
	if err := repository.LinkAttachmentToStory(ctx, fixture.storyA, attachment.ID, fixture.workspaceB); !errors.Is(err, attachmentdomain.ErrNotFound) {
		t.Fatalf("cross-workspace link error = %v", err)
	}
	if err := repository.LinkAttachmentToStory(ctx, fixture.storyA, attachment.ID, fixture.workspaceA); err != nil {
		t.Fatalf("link attachment: %v", err)
	}
	if err := repository.LinkAttachmentToStory(ctx, fixture.storyA, attachment.ID, fixture.workspaceA); err != nil {
		t.Fatalf("idempotent link attachment: %v", err)
	}
	if _, err := repository.AuthorizeStoryAttachment(ctx, fixture.storyA, attachment.ID, fixture.workspaceA); err != nil {
		t.Fatalf("authorize attachment: %v", err)
	}
	if _, err := repository.AuthorizeStoryAttachment(ctx, fixture.storyB, attachment.ID, fixture.workspaceB); !errors.Is(err, attachmentdomain.ErrNotFound) {
		t.Fatalf("cross-workspace authorization error = %v", err)
	}
	attachments, err := repository.GetAttachmentsByStoryID(ctx, fixture.storyA, fixture.workspaceA)
	if err != nil || len(attachments) != 1 || attachments[0].ID != attachment.ID {
		t.Fatalf("story attachments = %#v, %v", attachments, err)
	}
	if deleted, err := repository.DeleteAttachmentIfUnreferenced(ctx, attachment.ID, fixture.workspaceA); err != nil || deleted {
		t.Fatalf("referenced attachment deletion = %t, %v", deleted, err)
	}
	if err := repository.DeleteAttachment(ctx, attachment.ID, fixture.workspaceB); !errors.Is(err, attachmentdomain.ErrNotFound) {
		t.Fatalf("cross-workspace delete error = %v", err)
	}
}

func TestRepositoryAtomicallyUnlinksAndDeletesOrphanedStoryMedia(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	fixture := newAttachmentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	attachment := createFixtureAttachment(t, ctx, repository, fixture.workspaceA, fixture.userA, attachmentdomain.OptimizationNotRequested)

	if err := repository.LinkStoryMedia(ctx, fixture.storyA, attachment.ID, fixture.userB, fixture.workspaceA); !errors.Is(err, attachmentdomain.ErrNotFound) {
		t.Fatalf("cross-workspace creator link error = %v", err)
	}
	if err := repository.LinkStoryMedia(ctx, fixture.storyA, attachment.ID, fixture.userA, fixture.workspaceA); err != nil {
		t.Fatalf("link story media: %v", err)
	}
	if _, err := repository.AuthorizeStoryMedia(ctx, fixture.storyA, attachment.ID, fixture.workspaceB); !errors.Is(err, attachmentdomain.ErrNotFound) {
		t.Fatalf("cross-workspace media authorization error = %v", err)
	}
	if orphaned, err := repository.UnlinkStoryMedia(ctx, fixture.storyB, attachment.ID, fixture.workspaceB); !errors.Is(err, attachmentdomain.ErrNotFound) || orphaned {
		t.Fatalf("cross-workspace unlink = %t, %v", orphaned, err)
	}
	orphaned, err := repository.UnlinkStoryMedia(ctx, fixture.storyA, attachment.ID, fixture.workspaceA)
	if err != nil || !orphaned {
		t.Fatalf("unlink orphaned story media = %t, %v", orphaned, err)
	}
	if _, err := repository.GetAttachmentByID(ctx, attachment.ID, fixture.workspaceA); !errors.Is(err, attachmentdomain.ErrNotFound) {
		t.Fatalf("orphaned attachment remained: %v", err)
	}
}

func TestRepositoryFencesConcurrentOptimizationClaims(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	fixture := newAttachmentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	attachment := createFixtureAttachment(t, ctx, repository, fixture.workspaceA, fixture.userA, attachmentdomain.OptimizationQueued)

	start := make(chan struct{})
	errorsByClaim := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			_, err := repository.StartAttachmentOptimization(ctx, attachment.ID, fixture.workspaceA, time.Minute)
			errorsByClaim <- err
		}()
	}
	ready.Wait()
	close(start)

	var succeeded, conflicted int
	for range 2 {
		err := <-errorsByClaim
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, attachmentdomain.ErrStateConflict):
			conflicted++
		default:
			t.Fatalf("unexpected optimization claim error: %v", err)
		}
	}
	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("optimization claims succeeded=%d conflicted=%d", succeeded, conflicted)
	}
	if err := repository.CompleteAttachmentOptimization(
		ctx,
		attachment.ID,
		fixture.workspaceB,
		123,
		"image/jpeg",
		attachmentdomain.OptimizationSucceeded,
	); !errors.Is(err, attachmentdomain.ErrStateConflict) {
		t.Fatalf("cross-workspace completion error = %v", err)
	}
	if err := repository.CompleteAttachmentOptimization(
		ctx,
		attachment.ID,
		fixture.workspaceA,
		123,
		"image/jpeg",
		attachmentdomain.OptimizationSucceeded,
	); err != nil {
		t.Fatalf("complete optimization: %v", err)
	}
	stored, err := repository.GetAttachmentByID(ctx, attachment.ID, fixture.workspaceA)
	if err != nil || stored.OptimizationStatus != attachmentdomain.OptimizationSucceeded || stored.Size != 123 || stored.OptimizationAttempts != 1 {
		t.Fatalf("completed attachment = %#v, %v", stored, err)
	}
}

type attachmentFixture struct {
	workspaceA uuid.UUID
	workspaceB uuid.UUID
	userA      uuid.UUID
	userB      uuid.UUID
	storyA     uuid.UUID
	storyB     uuid.UUID
}

func newAttachmentFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) attachmentFixture {
	t.Helper()
	workspaceA := insertAttachmentWorkspace(t, ctx, pool, "a")
	workspaceB := insertAttachmentWorkspace(t, ctx, pool, "b")
	userA := insertAttachmentUser(t, ctx, pool, "a")
	userB := insertAttachmentUser(t, ctx, pool, "b")
	insertAttachmentMember(t, ctx, pool, workspaceA, userA)
	insertAttachmentMember(t, ctx, pool, workspaceB, userB)
	teamA := insertAttachmentTeam(t, ctx, pool, workspaceA, "A")
	teamB := insertAttachmentTeam(t, ctx, pool, workspaceB, "B")
	return attachmentFixture{
		workspaceA: workspaceA,
		workspaceB: workspaceB,
		userA:      userA,
		userB:      userB,
		storyA:     insertAttachmentStory(t, ctx, pool, workspaceA, teamA, userA, "A"),
		storyB:     insertAttachmentStory(t, ctx, pool, workspaceB, teamB, userB, "B"),
	}
}

func createFixtureAttachment(
	t *testing.T,
	ctx context.Context,
	repository *Repository,
	workspaceID, userID uuid.UUID,
	status attachmentdomain.OptimizationStatus,
) attachmentdomain.Attachment {
	t.Helper()
	attachment, err := repository.CreateAttachment(ctx, attachmentdomain.Attachment{
		Filename:           "evidence.jpg",
		BlobName:           uuid.NewString() + ".jpg",
		Size:               4096,
		MimeType:           "image/jpeg",
		UploadedBy:         userID,
		WorkspaceID:        workspaceID,
		ScanStatus:         attachmentdomain.ScanStatusUnscanned,
		OptimizationStatus: status,
	})
	if err != nil {
		t.Fatalf("create attachment: %v", err)
	}
	return attachment
}

func insertAttachmentWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`,
		id, "Attachment "+label, "attachment-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertAttachmentUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, TRUE, FALSE)
	`, id, "attachment-"+label+"-"+id.String(), fmt.Sprintf("attachment-%s-%s@example.com", label, id), "Attachment "+label); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertAttachmentMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST('member' AS user_role))
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertAttachmentTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, "Attachment "+label, workspaceID, "AT"+uuid.NewString()[:6]); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func insertAttachmentStory(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID, teamID, reporterID uuid.UUID,
	label string,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO stories (id, sequence_id, team_id, title, workspace_id, reporter_id)
		VALUES ($1, 1, $2, $3, $4, $5)
	`, id, teamID, "Attachment story "+label, workspaceID, reporterID); err != nil {
		t.Fatalf("insert story: %v", err)
	}
	return id
}
