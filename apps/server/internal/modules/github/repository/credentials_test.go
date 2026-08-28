package githubrepository

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

func TestNormalizedCredentialPageLimitIsBoundedAndTyped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		requested int
		want      int
	}{
		{name: "default for negative", requested: -1, want: credentialvault.DefaultMaintenanceBatchSize},
		{name: "default for zero", requested: 0, want: credentialvault.DefaultMaintenanceBatchSize},
		{name: "requested size", requested: 25, want: 25},
		{name: "maximum", requested: credentialvault.MaxMaintenanceBatchSize, want: credentialvault.MaxMaintenanceBatchSize},
		{name: "clamped", requested: credentialvault.MaxMaintenanceBatchSize + 1, want: credentialvault.MaxMaintenanceBatchSize},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			normalized, queryLimit, err := normalizedCredentialPageLimit(test.requested)
			if err != nil {
				t.Fatalf("normalizedCredentialPageLimit(%d): %v", test.requested, err)
			}
			if normalized != test.want || int(queryLimit) != test.want {
				t.Fatalf(
					"normalizedCredentialPageLimit(%d) = (%d, %d), want (%d, %d)",
					test.requested,
					normalized,
					queryLimit,
					test.want,
					test.want,
				)
			}
		})
	}
}

func TestCredentialWritesRejectPlaintextBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()
	repository := &Repo{}
	userID := uuid.New()
	generation := uuid.New()

	if err := repository.LinkGitHubUser(
		context.Background(),
		userID,
		42,
		"octocat",
		"gho_plaintext",
		credentialvault.CurrentVersion,
		generation,
	); err == nil {
		t.Fatal("LinkGitHubUser(plaintext) error = nil")
	}
	if err := repository.UpgradeLegacyGitHubUserCredential(
		context.Background(),
		userID,
		"gho_plaintext",
		"still-plaintext",
		credentialvault.CurrentVersion,
		generation,
	); err == nil {
		t.Fatal("UpgradeLegacyGitHubUserCredential(plaintext) error = nil")
	}
	if _, err := repository.RewrapGitHubUserCredential(
		context.Background(),
		GitHubUserCredentialRecord{
			UserID:          userID,
			Payload:         "vault.v2.expected",
			EnvelopeVersion: credentialvault.CurrentVersion,
			Generation:      generation,
		},
		"gho_rewrapped_plaintext",
	); err == nil {
		t.Fatal("RewrapGitHubUserCredential(plaintext) error = nil")
	}
}
