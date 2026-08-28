//go:build integration

package githubrepository

import (
	"errors"
	"testing"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestWebhookAuthorizationQueriesFenceInstallationGenerationOnPostgres18(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	ownerID := uuid.New()
	workspaceID := uuid.New()
	installationID := uuid.New()
	repositoryID := uuid.New()
	suffix := uuid.NewString()
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'GitHub webhook owner')
	`, ownerID, "github-webhook-"+suffix, "github-webhook-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert owner: %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'GitHub webhook workspace', $2, $3)
	`, workspaceID, "github-webhook-"+suffix, ownerID); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO github_installations (
			id, workspace_id, github_app_id, github_installation_id, account_id,
			account_login, account_type, repository_selection, permissions, events,
			installed_by_user_id, is_active
		) VALUES ($1, $2, 7, 41, 19, 'complexus', 'Organization', 'selected', '{}', '[]', $3, TRUE)
	`, installationID, workspaceID, ownerID); err != nil {
		t.Fatalf("insert installation: %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO github_repositories (
			id, workspace_id, installation_id, github_repository_id, owner_id,
			owner_login, name, full_name, html_url, clone_url, ssh_url, is_active
		) VALUES ($1, $2, $3, 42, 19, 'complexus', 'fortyone', 'complexus/fortyone',
			'https://github.com/complexus/fortyone', 'https://github.com/complexus/fortyone.git',
			'git@github.com:complexus/fortyone.git', TRUE)
	`, repositoryID, workspaceID, installationID); err != nil {
		t.Fatalf("insert repository: %v", err)
	}

	repository := New(postgres.Pool)
	authorized, err := repository.GetAuthorizedWebhookInstallation(ctx, 41, 42)
	if err != nil {
		t.Fatalf("GetAuthorizedWebhookInstallation() error = %v", err)
	}
	if authorized.ID != installationID || authorized.WorkspaceID != workspaceID || authorized.InstallationGeneration == uuid.Nil {
		t.Fatalf("authorized installation = %#v", authorized)
	}
	oldGeneration := authorized.InstallationGeneration
	if _, err := postgres.Pool.Exec(ctx, `UPDATE github_installations SET is_active = FALSE WHERE id = $1`, installationID); err != nil {
		t.Fatalf("deactivate installation: %v", err)
	}
	if _, err := repository.GetAuthorizedWebhookInstallation(ctx, 41, 42); !errors.Is(err, githubshared.ErrWebhookInstallationNotFound) {
		t.Fatalf("inactive authorization error = %v", err)
	}
	if _, err := repository.GetCurrentWebhookInstallation(ctx, installationID, oldGeneration, 42); !errors.Is(err, githubshared.ErrWebhookInstallationNotFound) {
		t.Fatalf("stale generation error = %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `UPDATE github_installations SET is_active = TRUE WHERE id = $1`, installationID); err != nil {
		t.Fatalf("reauthorize installation: %v", err)
	}
	current, err := repository.GetAuthorizedWebhookInstallation(ctx, 41, 42)
	if err != nil {
		t.Fatalf("GetAuthorizedWebhookInstallation(reauthorized) error = %v", err)
	}
	if current.InstallationGeneration == oldGeneration {
		t.Fatal("installation generation did not rotate across disconnect and reauthorization")
	}
	if _, err := repository.GetCurrentWebhookInstallation(ctx, installationID, current.InstallationGeneration, 42); err != nil {
		t.Fatalf("GetCurrentWebhookInstallation(current) error = %v", err)
	}
}
