package users

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

const testVerificationSecret = "test-verification-token-key-with-32-bytes"

func TestVerificationTokenManagerIssuesCSPRNGDigestOnlyInput(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	manager, err := newVerificationTokenManager(
		VerificationTokenConfig{Current: VerificationTokenKey{ID: "2026-08-v1", Secret: testVerificationSecret}},
		bytes.NewReader([]byte{0, 0, 42}),
		func() time.Time { return now },
	)
	if err != nil {
		t.Fatalf("construct verification token manager: %v", err)
	}

	code, input, err := manager.issue(" User@Example.COM ", TokenTypeRegistration, now.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("issue verification token: %v", err)
	}
	if code != "000042" {
		t.Fatalf("code = %q, want fixed six-digit code", code)
	}
	if input.Email != "user@example.com" {
		t.Fatalf("email = %q, want normalized address", input.Email)
	}
	if input.TokenType != TokenTypeRegistration {
		t.Fatalf("purpose = %q, want %q", input.TokenType, TokenTypeRegistration)
	}
	if len(input.TokenDigest) != 32 {
		t.Fatalf("digest length = %d, want 32", len(input.TokenDigest))
	}
	if input.TokenKeyID != "2026-08-v1" || input.TokenVersion != verificationTokenVersion {
		t.Fatalf("key metadata = (%q, %d), want versioned key", input.TokenKeyID, input.TokenVersion)
	}
	if input.MaximumIssues != verificationTokenIssueLimit {
		t.Fatalf("issue limit = %d, want %d", input.MaximumIssues, verificationTokenIssueLimit)
	}
}

func TestVerificationTokenDigestBindsEmailPurposeAndCode(t *testing.T) {
	t.Parallel()

	key := verificationTokenKey{id: "v1", secret: []byte(testVerificationSecret)}
	baseline := hex.EncodeToString(verificationTokenDigest(key, "user@example.com", TokenTypeRegistration, "123456"))
	variants := []string{
		hex.EncodeToString(verificationTokenDigest(key, "other@example.com", TokenTypeRegistration, "123456")),
		hex.EncodeToString(verificationTokenDigest(key, "user@example.com", TokenTypeLogin, "123456")),
		hex.EncodeToString(verificationTokenDigest(key, "user@example.com", TokenTypeRegistration, "654321")),
	}
	for _, variant := range variants {
		if variant == baseline {
			t.Fatal("verification digest did not bind every security context field")
		}
	}
}

func TestVerificationTokenConsumptionSupportsBoundedKeyRotation(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	manager, err := newVerificationTokenManager(
		VerificationTokenConfig{
			Current: VerificationTokenKey{ID: "v2", Secret: "current-verification-token-key-with-32-bytes"},
			Previous: []VerificationTokenKey{
				{ID: "v1", Secret: "previous-verification-token-key-with-32-bytes"},
			},
		},
		bytes.NewReader([]byte{0, 0, 1}),
		func() time.Time { return now },
	)
	if err != nil {
		t.Fatalf("construct verification token manager: %v", err)
	}

	input, err := manager.consumption("USER@example.com", "123456", TokenTypeRegistration, TokenTypeLogin)
	if err != nil {
		t.Fatalf("build consumption input: %v", err)
	}
	if len(input.TokenDigests) != 4 {
		t.Fatalf("digest candidates = %d, want current/previous x two purposes", len(input.TokenDigests))
	}
	if strings.Join(input.TokenKeyIDs, ",") != "v2,v2,v1,v1" {
		t.Fatalf("key IDs = %v, want one key ID paired with each digest candidate", input.TokenKeyIDs)
	}
	if len(input.TokenVersions) != len(input.TokenDigests) {
		t.Fatalf("versions = %v, want one version paired with each digest candidate", input.TokenVersions)
	}
	if input.Email != "user@example.com" || input.LegacyToken != "123456" {
		t.Fatalf("consumption input was not normalized: %#v", input)
	}
}

func TestVerificationRateLimitKeyDoesNotExposeIdentity(t *testing.T) {
	t.Parallel()

	manager, err := NewVerificationTokenManager(VerificationTokenConfig{
		Current: VerificationTokenKey{ID: "v1", Secret: testVerificationSecret},
	})
	if err != nil {
		t.Fatalf("construct verification token manager: %v", err)
	}

	key, err := manager.RateLimitKey("confirm", "email", "private@example.com")
	if err != nil {
		t.Fatalf("derive rate limit key: %v", err)
	}
	if strings.Contains(key, "private@example.com") {
		t.Fatalf("rate limit key leaked raw identity: %q", key)
	}
	if !strings.HasPrefix(key, "auth-verification:v1:v1:confirm:") {
		t.Fatalf("rate limit key = %q, want versioned prefix", key)
	}
}

func TestVerificationTokenManagerRejectsUnsafeKeysAndCodes(t *testing.T) {
	t.Parallel()

	_, err := NewVerificationTokenManager(VerificationTokenConfig{
		Current: VerificationTokenKey{ID: "v1", Secret: "short"},
	})
	if err == nil || !strings.Contains(err.Error(), "at least 32 bytes") {
		t.Fatalf("unsafe key error = %v", err)
	}

	manager, err := NewVerificationTokenManager(VerificationTokenConfig{
		Current: VerificationTokenKey{ID: "v1", Secret: testVerificationSecret},
	})
	if err != nil {
		t.Fatalf("construct verification token manager: %v", err)
	}
	if _, err := manager.consumption("user@example.com", "12345x", TokenTypeRegistration); err != ErrInvalidToken {
		t.Fatalf("invalid code error = %v, want %v", err, ErrInvalidToken)
	}
}
