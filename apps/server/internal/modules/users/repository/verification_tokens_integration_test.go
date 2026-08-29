//go:build integration

package usersrepository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

func TestVerificationTokenSecurityContract(t *testing.T) {
	postgres := testkit.NewPostgres(t)

	log := logger.NewWithText(io.Discard, slog.LevelError, "verification-token-integration-test")
	manager, err := users.NewVerificationTokenManager(users.VerificationTokenConfig{
		Current: users.VerificationTokenKey{
			ID:     "integration-v1",
			Secret: strings.Repeat("integration-test-key-", 2),
		},
		Previous: []users.VerificationTokenKey{{
			ID:     "integration-v0",
			Secret: strings.Repeat("previous-integration-key-", 2),
		}},
	})
	if err != nil {
		t.Fatalf("construct verification token manager: %v", err)
	}
	verificationTokens := New(postgres.Pool)
	service := users.New(log, nil, nil, users.WithVerificationTokens(manager, verificationTokens))

	t.Run("issuance quota is atomic and plaintext is absent", func(t *testing.T) {
		const requests = 12
		results := issueVerificationTokensConcurrently(
			t,
			service,
			"parallel@example.com",
			users.TokenTypeRegistration,
			requests,
		)

		var issued []users.CoreVerificationToken
		limited := 0
		for _, result := range results {
			switch {
			case result.err == nil:
				issued = append(issued, result.token)
			case errors.Is(result.err, users.ErrTooManyAttempts):
				limited++
			default:
				t.Fatalf("unexpected issue error: %v", result.err)
			}
		}
		if len(issued) != 3 || limited != requests-3 {
			t.Fatalf("issued=%d limited=%d, want 3/%d", len(issued), limited, requests-3)
		}

		var totalRows, plaintextRows, digestRows int
		if err := postgres.Pool.QueryRow(t.Context(), `
			SELECT
				COUNT(*),
				COUNT(token),
				COUNT(token_digest)
			FROM public.verification_tokens
			WHERE email = $1
		`, "parallel@example.com").Scan(&totalRows, &plaintextRows, &digestRows); err != nil {
			t.Fatalf("inspect stored verification tokens: %v", err)
		}
		if totalRows != 3 || plaintextRows != 0 || digestRows != 3 {
			t.Fatalf("rows/plaintext/digests = %d/%d/%d, want 3/0/3", totalRows, plaintextRows, digestRows)
		}

		consumeResults := consumeVerificationTokenConcurrently(
			t,
			service,
			"parallel@example.com",
			issued[0].Token,
			users.TokenTypeRegistration,
			requests,
		)
		succeeded := 0
		rejected := 0
		for _, consumeErr := range consumeResults {
			switch {
			case consumeErr == nil:
				succeeded++
			case errors.Is(consumeErr, users.ErrInvalidToken):
				rejected++
			default:
				t.Fatalf("unexpected consume error: %v", consumeErr)
			}
		}
		if succeeded != 1 || rejected != requests-1 {
			t.Fatalf("consume succeeded/rejected = %d/%d, want 1/%d", succeeded, rejected, requests-1)
		}
	})

	t.Run("digest is bound to email and purpose", func(t *testing.T) {
		issued, err := service.CreateVerificationToken(
			t.Context(),
			"purpose@example.com",
			users.TokenTypeRegistration,
			time.Now().Add(10*time.Minute),
		)
		if err != nil {
			t.Fatalf("issue purpose-bound token: %v", err)
		}

		if _, err := service.ConsumeVerificationToken(
			t.Context(),
			"other@example.com",
			issued.Token,
			users.TokenTypeRegistration,
		); !errors.Is(err, users.ErrInvalidToken) {
			t.Fatalf("cross-email consume error = %v, want invalid token", err)
		}
		if _, err := service.ConsumeVerificationToken(
			t.Context(),
			"purpose@example.com",
			issued.Token,
			users.TokenTypeLogin,
		); !errors.Is(err, users.ErrInvalidToken) {
			t.Fatalf("cross-purpose consume error = %v, want invalid token", err)
		}
		if _, err := service.ConsumeVerificationToken(
			t.Context(),
			"purpose@example.com",
			issued.Token,
			users.TokenTypeRegistration,
		); err != nil {
			t.Fatalf("consume with correct binding: %v", err)
		}
	})

	t.Run("digest metadata candidates remain tuple-bound", func(t *testing.T) {
		issued, err := service.CreateVerificationToken(
			t.Context(),
			"metadata-binding@example.com",
			users.TokenTypeRegistration,
			time.Now().Add(10*time.Minute),
		)
		if err != nil {
			t.Fatalf("issue metadata-bound token: %v", err)
		}

		if _, err := postgres.Pool.Exec(t.Context(), `
			UPDATE public.verification_tokens
			SET token_key_id = 'integration-v0'
			WHERE id = $1
		`, issued.ID); err != nil {
			t.Fatalf("mismatch persisted key metadata: %v", err)
		}

		if _, err := service.ConsumeVerificationToken(
			t.Context(),
			"metadata-binding@example.com",
			issued.Token,
			users.TokenTypeRegistration,
		); !errors.Is(err, users.ErrInvalidToken) {
			t.Fatalf("mismatched key metadata consume error = %v, want invalid token", err)
		}
	})

	t.Run("legacy plaintext row remains consumable during expand window", func(t *testing.T) {
		now := time.Now().UTC()
		if _, err := postgres.Pool.Exec(t.Context(), `
			INSERT INTO public.verification_tokens (
				token, email, expires_at, token_type, created_at, updated_at
			) VALUES ($1, $2, $3, CAST($4 AS public.token_type), $5, $5)
		`, "987654", "legacy@example.com", now.Add(10*time.Minute), users.TokenTypeRegistration, now); err != nil {
			t.Fatalf("insert legacy verification token: %v", err)
		}

		consumed, err := service.ConsumeVerificationToken(
			t.Context(),
			"legacy@example.com",
			"987654",
			users.TokenTypeRegistration,
		)
		if err != nil {
			t.Fatalf("consume legacy verification token: %v", err)
		}
		if consumed.TokenKeyID != "" || consumed.TokenVersion != 0 || consumed.UsedAt == nil {
			t.Fatalf("legacy token metadata = %#v, want consumed legacy row", consumed)
		}
	})
}

type verificationTokenIssueResult struct {
	token users.CoreVerificationToken
	err   error
}

func issueVerificationTokensConcurrently(
	t *testing.T,
	service *users.Service,
	email string,
	tokenType string,
	requests int,
) []verificationTokenIssueResult {
	t.Helper()

	start := make(chan struct{})
	results := make(chan verificationTokenIssueResult, requests)
	var waitGroup sync.WaitGroup
	for range requests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			token, err := service.CreateVerificationToken(
				context.Background(),
				email,
				tokenType,
				time.Now().Add(10*time.Minute),
			)
			results <- verificationTokenIssueResult{token: token, err: err}
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	collected := make([]verificationTokenIssueResult, 0, requests)
	for result := range results {
		collected = append(collected, result)
	}
	return collected
}

func consumeVerificationTokenConcurrently(
	t *testing.T,
	service *users.Service,
	email string,
	token string,
	tokenType string,
	requests int,
) []error {
	t.Helper()

	start := make(chan struct{})
	results := make(chan error, requests)
	var waitGroup sync.WaitGroup
	for range requests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := service.ConsumeVerificationToken(
				context.Background(),
				email,
				token,
				tokenType,
			)
			results <- err
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	collected := make([]error, 0, requests)
	for err := range results {
		collected = append(collected, err)
	}
	return collected
}
