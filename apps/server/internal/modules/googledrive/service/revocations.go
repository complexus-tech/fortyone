package googledrive

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

const (
	revocationBatchSize       = 25
	revocationLeaseDuration   = 2 * time.Minute
	revocationFinalizeTimeout = 5 * time.Second
	revocationMaximumAttempts = 8
	revocationInitialDelay    = 30 * time.Second
	revocationMaximumDelay    = 30 * time.Minute
)

type RevocationRepository interface {
	WithinProviderUserLifecycle(context.Context, uuid.UUID, func(context.Context) error) error
	WithinProviderSubjectLifecycle(context.Context, string, func(context.Context) error) error
	ListReadyRevocations(context.Context, time.Time, int) ([]domain.RevocationCandidate, error)
	ClaimRevocation(context.Context, domain.RevocationCandidate, uuid.UUID, time.Time, time.Time) (domain.Revocation, bool, error)
	CompleteRevocation(context.Context, uuid.UUID, uuid.UUID, time.Time) error
	RetryRevocation(context.Context, uuid.UUID, uuid.UUID, string, time.Time, time.Time, bool) error
}

type RevocationDispatcher struct {
	log         *logger.Logger
	repo        RevocationRepository
	credentials CredentialVault
	client      ProviderClient
	now         func() time.Time
}

func NewRevocationDispatcher(
	log *logger.Logger,
	repo RevocationRepository,
	credentials CredentialVault,
) *RevocationDispatcher {
	return &RevocationDispatcher{
		log: log, repo: repo, credentials: credentials,
		client: newGoogleClient(&http.Client{Timeout: 25 * time.Second}, Config{}, nil),
		now:    time.Now,
	}
}

func (dispatcher *RevocationDispatcher) DispatchPendingRevocations(ctx context.Context) (int, error) {
	if dispatcher == nil || dispatcher.repo == nil || dispatcher.credentials == nil ||
		dispatcher.client == nil || dispatcher.now == nil {
		return 0, errors.New("Google Drive revocation dispatcher is not configured")
	}
	candidates, err := dispatcher.repo.ListReadyRevocations(ctx, dispatcher.now().UTC(), revocationBatchSize)
	if err != nil {
		return 0, fmt.Errorf("list Google Drive revocations: %w", err)
	}

	var dispatchErr error
	for _, candidate := range candidates {
		if err := dispatcher.dispatch(ctx, candidate); err != nil {
			dispatchErr = errors.Join(dispatchErr, fmt.Errorf("dispatch Google Drive revocation %s: %w", candidate.ID, err))
		}
	}
	return len(candidates), dispatchErr
}

func (dispatcher *RevocationDispatcher) dispatch(
	ctx context.Context,
	candidate domain.RevocationCandidate,
) error {
	return dispatcher.repo.WithinProviderUserLifecycle(ctx, candidate.UserID, func(userCtx context.Context) error {
		return dispatcher.repo.WithinProviderSubjectLifecycle(userCtx, candidate.GoogleSubject, func(subjectCtx context.Context) error {
			claimedAt := dispatcher.now().UTC()
			revocation, claimed, err := dispatcher.repo.ClaimRevocation(
				subjectCtx,
				candidate,
				uuid.New(),
				claimedAt,
				claimedAt.Add(revocationLeaseDuration),
			)
			if err != nil || !claimed {
				return err
			}

			token, err := openRevocationToken(dispatcher.credentials, revocation.Account())
			if err != nil {
				return dispatcher.release(
					subjectCtx,
					revocation,
					"Google Drive revocation credential could not be opened",
					true,
				)
			}
			revokeErr := dispatcher.client.Revoke(subjectCtx, revocationToken(token))
			if revokeErr != nil && !revocationAlreadySatisfied(revokeErr) {
				terminal := revocation.AttemptCount >= revocationMaximumAttempts || permanentRevocationError(revokeErr)
				failure := safeRevocationFailure(revokeErr)
				return errors.Join(
					errors.New(failure),
					dispatcher.release(subjectCtx, revocation, failure, terminal),
				)
			}

			finalizeCtx, cancel := context.WithTimeout(context.WithoutCancel(subjectCtx), revocationFinalizeTimeout)
			defer cancel()
			if err := dispatcher.repo.CompleteRevocation(
				finalizeCtx,
				revocation.ID,
				revocation.ClaimToken,
				dispatcher.now().UTC(),
			); err != nil {
				return fmt.Errorf("complete remote revocation: %w", err)
			}
			return nil
		})
	})
}

func (dispatcher *RevocationDispatcher) release(
	ctx context.Context,
	revocation domain.Revocation,
	failure string,
	terminal bool,
) error {
	releasedAt := dispatcher.now().UTC()
	retryAt := releasedAt.Add(revocationRetryDelay(revocation.AttemptCount))
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), revocationFinalizeTimeout)
	defer cancel()
	if err := dispatcher.repo.RetryRevocation(
		releaseCtx,
		revocation.ID,
		revocation.ClaimToken,
		failure,
		retryAt,
		releasedAt,
		terminal,
	); err != nil {
		return fmt.Errorf("release remote revocation claim: %w", err)
	}
	if dispatcher.log != nil {
		dispatcher.log.Warn(
			releaseCtx,
			"Google Drive remote revocation failed",
			"revocation_id", revocation.ID,
			"attempt", revocation.AttemptCount,
			"terminal", terminal,
			"error", failure,
		)
	}
	return nil
}

func safeRevocationFailure(err error) string {
	var apiError *APIError
	if errors.As(err, &apiError) {
		switch {
		case apiError.IsRateLimited():
			return "Google Drive remote revocation was rate limited"
		case apiError.StatusCode >= http.StatusInternalServerError:
			return "Google Drive remote revocation service was unavailable"
		default:
			return fmt.Sprintf("Google Drive remote revocation was rejected with status %d", apiError.StatusCode)
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "Google Drive remote revocation timed out"
	}
	if errors.Is(err, context.Canceled) {
		return "Google Drive remote revocation was canceled"
	}
	return "Google Drive remote revocation request failed"
}

func revocationAlreadySatisfied(err error) bool {
	var apiError *APIError
	return errors.As(err, &apiError) &&
		apiError.StatusCode == http.StatusBadRequest && apiError.Code == "invalid_token"
}

func permanentRevocationError(err error) bool {
	var apiError *APIError
	if !errors.As(err, &apiError) {
		return false
	}
	if apiError.IsRateLimited() {
		return false
	}
	if apiError.StatusCode < 400 || apiError.StatusCode >= 500 {
		return false
	}
	switch apiError.StatusCode {
	case http.StatusRequestTimeout, http.StatusConflict, http.StatusTooManyRequests:
		return false
	default:
		return true
	}
}

func revocationRetryDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return revocationInitialDelay
	}
	delay := revocationInitialDelay
	for current := 1; current < attempt && delay < revocationMaximumDelay; current++ {
		delay *= 2
		if delay >= revocationMaximumDelay {
			return revocationMaximumDelay
		}
	}
	return delay
}
