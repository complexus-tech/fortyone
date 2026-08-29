//go:build integration

package githubrepository

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInstallationUpsertFencesOwnershipAndReconcilesRepositoryAccess(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	ownerID := seedGitHubRepositoryUser(t, postgres.Pool, "installation-owner")
	otherOwnerID := seedGitHubRepositoryUser(t, postgres.Pool, "installation-other")
	workspaceID := seedGitHubRepositoryWorkspace(t, postgres.Pool, ownerID, "installation-primary")
	otherWorkspaceID := seedGitHubRepositoryWorkspace(t, postgres.Pool, otherOwnerID, "installation-other")
	repository := New(postgres.Pool)

	installation := githubInstallationFixture(91001)
	repositories := []GithubRepositoryPayload{
		githubRepositoryFixture(92001, "fortyone"),
		githubRepositoryFixture(92002, "docs"),
	}
	if err := repository.UpsertInstallationWithRepositories(
		ctx, workspaceID, ownerID, 90001, installation, repositories,
	); err != nil {
		t.Fatalf("UpsertInstallationWithRepositories() error = %v", err)
	}

	if err := repository.UpsertInstallationWithRepositories(
		ctx, otherWorkspaceID, otherOwnerID, 90001, installation, repositories,
	); !errors.Is(err, ErrInstallationOwnershipConflict) {
		t.Fatalf("cross-workspace upsert error = %v, want ErrInstallationOwnershipConflict", err)
	}
	var persistedWorkspaceID uuid.UUID
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT workspace_id
		FROM github_installations
		WHERE github_installation_id = $1
	`, installation.ID).Scan(&persistedWorkspaceID); err != nil {
		t.Fatalf("read persisted installation owner: %v", err)
	}
	if persistedWorkspaceID != workspaceID {
		t.Fatalf("installation workspace = %s, want %s", persistedWorkspaceID, workspaceID)
	}

	repositories[0].Description = pointerTo("updated repository")
	if err := repository.UpsertInstallationWithRepositories(
		ctx, workspaceID, ownerID, 90001, installation, repositories[:1],
	); err != nil {
		t.Fatalf("reconcile selected repositories: %v", err)
	}
	assertGitHubRepositoryActivity(t, postgres.Pool, installation.ID, map[int64]bool{
		92001: true,
		92002: false,
	})

	if err := repository.UpsertInstallationWithRepositories(
		ctx, workspaceID, ownerID, 90001, installation, nil,
	); err != nil {
		t.Fatalf("reconcile empty repository selection: %v", err)
	}
	assertGitHubRepositoryActivity(t, postgres.Pool, installation.ID, map[int64]bool{
		92001: false,
		92002: false,
	})
}

func TestNativeStoryLinkLabelAndCommentWritesAreIdempotent(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	ownerID := seedGitHubRepositoryUser(t, postgres.Pool, "story-owner")
	workspaceID := seedGitHubRepositoryWorkspace(t, postgres.Pool, ownerID, "story-workspace")
	teamID, statusID, storyID := seedGitHubStory(t, postgres.Pool, workspaceID, ownerID)
	repository := New(postgres.Pool)
	installation := githubInstallationFixture(93001)
	if err := repository.UpsertInstallationWithRepositories(
		ctx,
		workspaceID,
		ownerID,
		90001,
		installation,
		[]GithubRepositoryPayload{githubRepositoryFixture(94001, "fortyone")},
	); err != nil {
		t.Fatalf("seed GitHub installation: %v", err)
	}
	repositoryRecord, err := repository.FindRepositoryByExternalID(ctx, 94001)
	if err != nil {
		t.Fatalf("FindRepositoryByExternalID() error = %v", err)
	}

	metadata := map[string]any{"source": "integration"}
	if err := repository.UpsertStoryLink(
		ctx, workspaceID, storyID, repositoryRecord.ID, "issue", 95001, 17, nil,
		"https://github.example/issues/17", "Initial title", "open", metadata,
	); err != nil {
		t.Fatalf("create story link: %v", err)
	}
	if err := repository.UpsertStoryLink(
		ctx, workspaceID, storyID, repositoryRecord.ID, "issue", 95001, 17, nil,
		"https://github.example/issues/17", "Updated title", "closed", metadata,
	); err != nil {
		t.Fatalf("update story link: %v", err)
	}
	var linkCount int
	var title, state string
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*), MAX(title), MAX(state)
		FROM github_story_links
		WHERE story_id = $1 AND repository_id = $2 AND external_type = 'issue' AND github_id = 95001
	`, storyID, repositoryRecord.ID).Scan(&linkCount, &title, &state); err != nil {
		t.Fatalf("read story link upsert result: %v", err)
	}
	if linkCount != 1 || title != "Updated title" || state != "closed" {
		t.Fatalf("story link result = count %d title %q state %q", linkCount, title, state)
	}

	const externalURL = "https://github.example/issues/17"
	linkTitle := "GitHub issue 17"
	runConcurrently(t, 16, func() error {
		return repository.EnsureStoryLink(ctx, storyID, &linkTitle, externalURL)
	})
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM story_links
		WHERE story_id = $1 AND url = $2
	`, storyID, externalURL).Scan(&linkCount); err != nil {
		t.Fatalf("count generic story links: %v", err)
	}
	if linkCount != 1 {
		t.Fatalf("generic story link count = %d, want 1", linkCount)
	}

	var labelIDs sync.Map
	runConcurrently(t, 16, func() error {
		ids, err := repository.ResolveOrCreateLabelsByName(ctx, workspaceID, teamID, []string{"Backend"})
		if err != nil {
			return err
		}
		if len(ids) != 1 {
			return fmt.Errorf("resolved label count = %d, want 1", len(ids))
		}
		labelIDs.Store(ids[0], struct{}{})
		return nil
	})
	uniqueLabels := 0
	labelIDs.Range(func(_, _ any) bool {
		uniqueLabels++
		return true
	})
	if uniqueLabels != 1 {
		t.Fatalf("concurrent label IDs = %d, want 1", uniqueLabels)
	}
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM labels
		WHERE workspace_id = $1 AND LOWER(name) = 'backend'
	`, workspaceID).Scan(&linkCount); err != nil {
		t.Fatalf("count GitHub-created labels: %v", err)
	}
	if linkCount != 1 {
		t.Fatalf("GitHub-created label count = %d, want 1", linkCount)
	}

	reserved, err := repository.ReserveInboundGitHubComment(
		ctx, workspaceID, storyID, repositoryRecord.ID, 96001, ownerID,
	)
	if err != nil || !reserved {
		t.Fatalf("first comment reservation = (%t, %v), want true", reserved, err)
	}
	reserved, err = repository.ReserveInboundGitHubComment(
		ctx, workspaceID, storyID, repositoryRecord.ID, 96001, ownerID,
	)
	if err != nil || reserved {
		t.Fatalf("duplicate comment reservation = (%t, %v), want false", reserved, err)
	}

	if category, err := repository.GetStatusCategory(ctx, statusID); err != nil || category != "started" {
		t.Fatalf("GetStatusCategory() = (%q, %v), want started", category, err)
	}
}

func seedGitHubRepositoryUser(t testing.TB, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'GitHub repository integration')
	`, id, "github-repository-"+label+"-"+suffix, "github-repository-"+label+"-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert GitHub repository user: %v", err)
	}
	return id
}

func seedGitHubRepositoryWorkspace(
	t testing.TB,
	pool *pgxpool.Pool,
	ownerID uuid.UUID,
	label string,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, id, "GitHub "+label, "github-"+label+"-"+uuid.NewString(), ownerID); err != nil {
		t.Fatalf("insert GitHub repository workspace: %v", err)
	}
	return id
}

func seedGitHubStory(
	t testing.TB,
	pool *pgxpool.Pool,
	workspaceID, reporterID uuid.UUID,
) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	teamID, statusID, storyID := uuid.New(), uuid.New(), uuid.New()
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, 'GitHub integration team', $2, $3, '#000000')
	`, teamID, workspaceID, "GH"+uuid.NewString()[:6]); err != nil {
		t.Fatalf("insert GitHub integration team: %v", err)
	}
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO statuses (status_id, name, category, workspace_id, team_id, color)
		VALUES ($1, 'Started', 'started', $2, $3, '#000000')
	`, statusID, workspaceID, teamID); err != nil {
		t.Fatalf("insert GitHub integration status: %v", err)
	}
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO stories (id, workspace_id, team_id, status_id, reporter_id, sequence_id, title)
		VALUES ($1, $2, $3, $4, $5, 17, 'GitHub integration story')
	`, storyID, workspaceID, teamID, statusID, reporterID); err != nil {
		t.Fatalf("insert GitHub integration story: %v", err)
	}
	return teamID, statusID, storyID
}

func githubInstallationFixture(id int64) GithubInstallationPayload {
	return GithubInstallationPayload{
		ID: id,
		Account: GithubInstallationAccountPayload{
			ID:    id + 1,
			Login: "complexus",
			Type:  "Organization",
		},
		RepositorySelection: "selected",
		Permissions:         map[string]string{"issues": "write"},
		Events:              []string{"issues"},
		Sender:              GithubInstallationSenderPayload{ID: id + 2},
	}
}

func githubRepositoryFixture(id int64, name string) GithubRepositoryPayload {
	return GithubRepositoryPayload{
		ID:            id,
		Name:          name,
		FullName:      "complexus/" + name,
		HTMLURL:       "https://github.example/complexus/" + name,
		CloneURL:      "https://github.example/complexus/" + name + ".git",
		SSHURL:        "git@github.example:complexus/" + name + ".git",
		DefaultBranch: "main",
		Owner:         GithubRepositoryOwnerPayload{ID: id + 1, Login: "complexus"},
	}
}

func assertGitHubRepositoryActivity(
	t testing.TB,
	pool *pgxpool.Pool,
	installationExternalID int64,
	want map[int64]bool,
) {
	t.Helper()
	rows, err := pool.Query(t.Context(), `
		SELECT repository.github_repository_id, repository.is_active
		FROM github_repositories AS repository
		INNER JOIN github_installations AS installation ON installation.id = repository.installation_id
		WHERE installation.github_installation_id = $1
	`, installationExternalID)
	if err != nil {
		t.Fatalf("list GitHub repository activity: %v", err)
	}
	defer rows.Close()
	got := make(map[int64]bool, len(want))
	for rows.Next() {
		var id int64
		var active bool
		if err := rows.Scan(&id, &active); err != nil {
			t.Fatalf("scan GitHub repository activity: %v", err)
		}
		got[id] = active
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate GitHub repository activity: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("repository activity rows = %d, want %d", len(got), len(want))
	}
	for id, active := range want {
		if got[id] != active {
			t.Fatalf("repository %d active = %t, want %t", id, got[id], active)
		}
	}
}

func runConcurrently(t testing.TB, count int, operation func() error) {
	t.Helper()
	errorsChannel := make(chan error, count)
	var waitGroup sync.WaitGroup
	for range count {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			errorsChannel <- operation()
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent repository operation: %v", err)
		}
	}
}

func pointerTo[T any](value T) *T {
	return &value
}
