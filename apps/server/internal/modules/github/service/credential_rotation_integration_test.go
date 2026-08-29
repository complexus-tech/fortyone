//go:build integration

package github

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"

	githubrepository "github.com/complexus-tech/projects-api/internal/modules/github/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGitHubCredentialRewrapPreservesGenerationAndFencesRelinkAndRevoke(t *testing.T) {
	database := testkit.NewPostgres(t)
	pool := database.Pool

	ctx := t.Context()
	repository := githubrepository.New(database.Pool)
	oldRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 1}
	newRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 2}
	oldKey := credentialvault.Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x31}, 32)}
	newKey := credentialvault.Key{Ref: newRef, Material: bytes.Repeat([]byte{0x42}, 32)}
	oldVault := newGitHubVault(t, oldRef, oldKey)
	rotatedVault := newGitHubVault(t, newRef, oldKey, newKey)

	oldUserID := seedGitHubCredentialUser(t, pool, "old")
	currentUserID := seedGitHubCredentialUser(t, pool, "current")
	oldGeneration := uuid.New()
	currentGeneration := uuid.New()
	oldEnvelope, err := oldVault.Seal(githubUserCredentialContext(oldUserID, oldGeneration), []byte("gho-old"))
	if err != nil {
		t.Fatalf("seal old GitHub credential: %v", err)
	}
	currentEnvelope, err := rotatedVault.Seal(githubUserCredentialContext(currentUserID, currentGeneration), []byte("gho-current"))
	if err != nil {
		t.Fatalf("seal current GitHub credential: %v", err)
	}
	if err := repository.LinkGitHubUser(ctx, oldUserID, 101, "old-user", oldEnvelope, credentialvault.CurrentVersion, oldGeneration); err != nil {
		t.Fatalf("persist old GitHub credential: %v", err)
	}
	if err := repository.LinkGitHubUser(ctx, currentUserID, 102, "current-user", currentEnvelope, credentialvault.CurrentVersion, currentGeneration); err != nil {
		t.Fatalf("persist current GitHub credential: %v", err)
	}
	service, err := New(nil, repository, nil, nil, nil, Config{CredentialVault: rotatedVault})
	if err != nil {
		t.Fatalf("construct GitHub service: %v", err)
	}

	report, err := service.RewrapUserCredentials(ctx)
	if err != nil {
		t.Fatalf("RewrapUserCredentials() error = %v", err)
	}
	if report.ActiveKey != newRef || report.Scanned != 2 || report.Rewrapped != 1 || report.Current != 1 || report.Stale != 0 {
		t.Fatalf("RewrapUserCredentials() = %#v", report)
	}
	oldStored, err := repository.GetUserGitHubCredential(ctx, oldUserID)
	if err != nil {
		t.Fatalf("read rewrapped GitHub credential: %v", err)
	}
	if oldStored.Generation != oldGeneration {
		t.Fatalf("rewrap changed credential generation: got %s want %s", oldStored.Generation, oldGeneration)
	}
	metadata, err := credentialvault.Inspect(oldStored.Payload)
	if err != nil || metadata.Key != newRef {
		t.Fatalf("rewrapped metadata = (%#v, %v)", metadata, err)
	}
	opened, err := rotatedVault.Open(githubUserCredentialContext(oldUserID, oldGeneration), oldStored.Payload)
	if err != nil {
		t.Fatalf("open rewrapped GitHub credential: %v", err)
	}
	defer opened.Destroy()
	plaintext := opened.Reveal()
	defer clear(plaintext)
	if string(plaintext) != "gho-old" {
		t.Fatal("rewrapped GitHub credential changed")
	}
	second, err := service.RewrapUserCredentials(ctx)
	if err != nil || second.Scanned != 2 || second.Current != 2 || second.Rewrapped != 0 || second.Stale != 0 {
		t.Fatalf("second RewrapUserCredentials() = (%#v, %v)", second, err)
	}

	// Read an old candidate, then simulate a provider relink before the
	// maintenance compare-and-swap. The new credential and generation win.
	raceUserID := seedGitHubCredentialUser(t, pool, "race")
	raceGeneration := uuid.New()
	raceEnvelope, err := oldVault.Seal(githubUserCredentialContext(raceUserID, raceGeneration), []byte("gho-race-old"))
	if err != nil {
		t.Fatalf("seal race credential: %v", err)
	}
	if err := repository.LinkGitHubUser(ctx, raceUserID, 103, "race-old", raceEnvelope, credentialvault.CurrentVersion, raceGeneration); err != nil {
		t.Fatalf("persist race credential: %v", err)
	}
	candidates, err := repository.ListGitHubUserCredentialsForRewrap(ctx, nil, credentialvault.MaxMaintenanceBatchSize)
	if err != nil {
		t.Fatalf("list race candidates: %v", err)
	}
	var candidate githubrepository.GitHubUserCredentialRecord
	for _, record := range candidates {
		if record.UserID == raceUserID {
			candidate = record
			break
		}
	}
	if candidate.UserID == uuid.Nil {
		t.Fatal("race credential was not listed for rewrap")
	}
	staleResult, err := rotatedVault.Rewrap(githubUserCredentialContext(raceUserID, raceGeneration), candidate.Payload)
	if err != nil || !staleResult.Changed {
		t.Fatalf("prepare stale rewrap: changed=%t err=%v", staleResult.Changed, err)
	}
	refreshedGeneration := uuid.New()
	refreshedEnvelope, err := rotatedVault.Seal(githubUserCredentialContext(raceUserID, refreshedGeneration), []byte("gho-race-new"))
	if err != nil {
		t.Fatalf("seal refreshed credential: %v", err)
	}
	if err := repository.LinkGitHubUser(ctx, raceUserID, 104, "race-new", refreshedEnvelope, credentialvault.CurrentVersion, refreshedGeneration); err != nil {
		t.Fatalf("persist refreshed credential: %v", err)
	}
	replaced, err := repository.RewrapGitHubUserCredential(ctx, candidate, staleResult.Envelope)
	if err != nil || replaced {
		t.Fatalf("stale RewrapGitHubUserCredential() = (%v, %v), want false", replaced, err)
	}
	final, err := repository.GetUserGitHubCredential(ctx, raceUserID)
	if err != nil || final.Generation != refreshedGeneration || final.Payload != refreshedEnvelope {
		t.Fatalf("concurrent relink was overwritten: err=%v", err)
	}

	// Local revoke clears both ciphertext and generation. A stale maintenance
	// write cannot resurrect the credential after the revoke.
	if err := repository.UnlinkGitHubUser(ctx, raceUserID); err != nil {
		t.Fatalf("revoke GitHub user credential: %v", err)
	}
	replaced, err = repository.RewrapGitHubUserCredential(ctx, final, refreshedEnvelope)
	if err != nil || replaced {
		t.Fatalf("rewrap after revoke = (%v, %v), want false", replaced, err)
	}
	if _, err := repository.GetUserGitHubCredential(ctx, raceUserID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("revoked credential lookup error = %v, want sql.ErrNoRows", err)
	}
}

func TestGitHubCredentialRewrapRejectsTamperingWithoutPersistence(t *testing.T) {
	database := testkit.NewPostgres(t)
	pool := database.Pool

	ctx := t.Context()
	repository := githubrepository.New(database.Pool)
	oldRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 1}
	newRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 2}
	oldKey := credentialvault.Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x51}, 32)}
	newKey := credentialvault.Key{Ref: newRef, Material: bytes.Repeat([]byte{0x62}, 32)}
	oldVault := newGitHubVault(t, oldRef, oldKey)
	rotatedVault := newGitHubVault(t, newRef, oldKey, newKey)
	userID := seedGitHubCredentialUser(t, pool, "tampered")
	generation := uuid.New()
	envelope, err := oldVault.Seal(githubUserCredentialContext(userID, generation), []byte("gho-tampered"))
	if err != nil {
		t.Fatalf("seal GitHub credential: %v", err)
	}
	tampered := envelope[:len(envelope)-1] + "A"
	if tampered == envelope {
		tampered = envelope[:len(envelope)-1] + "B"
	}
	if err := repository.LinkGitHubUser(ctx, userID, 105, "tampered", tampered, credentialvault.CurrentVersion, generation); err != nil {
		t.Fatalf("persist tampered fixture: %v", err)
	}
	service, err := New(nil, repository, nil, nil, nil, Config{CredentialVault: rotatedVault})
	if err != nil {
		t.Fatalf("construct GitHub service: %v", err)
	}

	if _, err := service.RewrapUserCredentials(ctx); err == nil {
		t.Fatal("RewrapUserCredentials(tampered) error = nil")
	}
	stored, err := repository.GetUserGitHubCredential(ctx, userID)
	if err != nil {
		t.Fatalf("read tampered fixture: %v", err)
	}
	if stored.Payload != tampered || stored.Generation != generation {
		t.Fatal("failed rewrap changed the tampered database record")
	}
}

func seedGitHubCredentialUser(t testing.TB, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'GitHub credential rotation')
	`, userID, "github-rotation-"+label+"-"+suffix, "github-rotation-"+label+"-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert GitHub credential user: %v", err)
	}
	return userID
}

func newGitHubVault(t testing.TB, active credentialvault.KeyRef, keys ...credentialvault.Key) *credentialvault.Vault {
	t.Helper()
	vault, err := credentialvault.New(credentialvault.Config{Active: active, Keys: keys})
	if err != nil {
		t.Fatalf("create GitHub credential vault: %v", err)
	}
	return vault
}
