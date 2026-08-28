//go:build integration

package github

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"
	"testing"

	githubrepository "github.com/complexus-tech/projects-api/internal/modules/github/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestGitHubCredentialBackfillEncryptsAndBindsLegacyTokens(t *testing.T) {
	database := testkit.NewPostgres(t)
	pool := database.Pool

	ctx := t.Context()
	userID := uuid.New()
	const legacyToken = "gho_legacy_token_must_not_survive"
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, github_access_token)
		VALUES ($1, $2, $3, 'GitHub credential migration', $4)
	`, userID, "github-vault-"+uuid.NewString(), "github-vault-"+uuid.NewString()+"@example.com", legacyToken); err != nil {
		t.Fatalf("insert legacy GitHub credential: %v", err)
	}

	vault := newGitHubIntegrationVault(t)
	repository := githubrepository.New(database.Pool)
	service, err := New(nil, repository, nil, nil, nil, Config{CredentialVault: vault})
	if err != nil {
		t.Fatalf("construct GitHub service: %v", err)
	}

	updated, err := service.BackfillLegacyUserCredentials(ctx)
	if err != nil {
		t.Fatalf("BackfillLegacyUserCredentials() error = %v", err)
	}
	if updated != 1 {
		t.Fatalf("BackfillLegacyUserCredentials() = %d, want 1", updated)
	}

	var stored string
	var version int
	var generation uuid.UUID
	if err := pool.QueryRow(ctx, `
		SELECT github_access_token, github_access_token_envelope_version, github_access_token_generation
		FROM users
		WHERE user_id = $1
	`, userID).Scan(&stored, &version, &generation); err != nil {
		t.Fatalf("read migrated GitHub credential: %v", err)
	}
	if !credentialvault.IsEnvelope(stored) || strings.Contains(stored, legacyToken) {
		t.Fatal("migrated GitHub credential is not a vault envelope")
	}
	if version != credentialvault.CurrentVersion || generation == uuid.Nil {
		t.Fatalf("migrated metadata = version %d, generation %s", version, generation)
	}
	opened, err := vault.Open(githubUserCredentialContext(userID, generation), stored)
	if err != nil {
		t.Fatalf("open migrated GitHub credential: %v", err)
	}
	defer opened.Destroy()
	plaintext := opened.Reveal()
	defer clear(plaintext)
	if !bytes.Equal(plaintext, []byte(legacyToken)) {
		t.Fatal("migrated GitHub credential changed")
	}
	if _, err := vault.Open(githubUserCredentialContext(uuid.New(), generation), stored); !errors.Is(err, credentialvault.ErrAuthentication) {
		t.Fatalf("open migrated credential under another user error = %v, want authentication failure", err)
	}
	if token, err := service.userGitHubToken(ctx, userID); err != nil || token != legacyToken {
		t.Fatalf("userGitHubToken() returned different data: err=%v", err)
	}
	if updated, err := service.BackfillLegacyUserCredentials(ctx); err != nil || updated != 0 {
		t.Fatalf("second BackfillLegacyUserCredentials() = (%d, %v), want no-op", updated, err)
	}

	if err := repository.UnlinkGitHubUser(ctx, userID); err != nil {
		t.Fatalf("UnlinkGitHubUser() error = %v", err)
	}
	var token *string
	var clearedVersion int
	var clearedGeneration *uuid.UUID
	if err := pool.QueryRow(ctx, `
		SELECT github_access_token, github_access_token_envelope_version, github_access_token_generation
		FROM users WHERE user_id = $1
	`, userID).Scan(&token, &clearedVersion, &clearedGeneration); err != nil {
		t.Fatalf("read unlinked GitHub credential: %v", err)
	}
	if token != nil || clearedVersion != 0 || clearedGeneration != nil {
		t.Fatalf("unlink retained credential metadata: token=%v version=%d generation=%v", token != nil, clearedVersion, clearedGeneration)
	}
}

func TestGitHubCredentialUpgradeCompareAndSwapDoesNotOverwriteConcurrentRelink(t *testing.T) {
	database := testkit.NewPostgres(t)
	pool := database.Pool

	ctx := t.Context()
	userID := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, github_access_token)
		VALUES ($1, $2, $3, 'old-plaintext')
	`, userID, "github-vault-cas-"+uuid.NewString(), "github-vault-cas-"+uuid.NewString()+"@example.com"); err != nil {
		t.Fatalf("insert legacy GitHub credential: %v", err)
	}

	repository := githubrepository.New(database.Pool)
	vault := newGitHubIntegrationVault(t)
	generation := uuid.New()
	encrypted, err := vault.Seal(githubUserCredentialContext(userID, generation), []byte("old-plaintext"))
	if err != nil {
		t.Fatalf("seal GitHub credential: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE users SET github_access_token = 'concurrent-plaintext' WHERE user_id = $1`, userID); err != nil {
		t.Fatalf("simulate concurrent GitHub relink: %v", err)
	}
	err = repository.UpgradeLegacyGitHubUserCredential(
		ctx,
		userID,
		"old-plaintext",
		encrypted,
		credentialvault.CurrentVersion,
		generation,
	)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("UpgradeLegacyGitHubUserCredential() error = %v, want sql.ErrNoRows", err)
	}
	var stored string
	var version int
	if err := pool.QueryRow(ctx, `
		SELECT github_access_token, github_access_token_envelope_version
		FROM users WHERE user_id = $1
	`, userID).Scan(&stored, &version); err != nil {
		t.Fatalf("read concurrent GitHub credential: %v", err)
	}
	if stored != "concurrent-plaintext" || version != 0 {
		t.Fatalf("CAS overwrote concurrent credential: version=%d", version)
	}
}

func newGitHubIntegrationVault(t testing.TB) *credentialvault.Vault {
	t.Helper()
	ref := credentialvault.KeyRef{ID: "integration", Version: 1}
	vault, err := credentialvault.New(credentialvault.Config{
		Active: ref,
		Keys: []credentialvault.Key{{
			Ref:      ref,
			Material: bytes.Repeat([]byte{0x7c}, 32),
		}},
	})
	if err != nil {
		t.Fatalf("create GitHub integration credential vault: %v", err)
	}
	return vault
}
