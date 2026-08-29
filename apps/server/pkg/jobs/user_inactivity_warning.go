package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	userInactivityWarningBatchSize  = 100
	userInactivityWarningMaxBatches = 100
	userInactivityWarningBatchDelay = 100 * time.Millisecond
)

var errUserInactivityWarningBacklogRemaining = errors.New("user inactivity warning backlog remains")

// UserInactivityWarningStore is the worker-owned account capability needed to
// page, recheck, and record lifecycle warning delivery.
type UserInactivityWarningStore interface {
	ListUserInactivityWarningCandidates(
		context.Context,
		usersdomain.InactivityWarningQuery,
	) ([]usersdomain.InactivityWarningCandidate, error)
	GetEligibleUserInactivityWarningCandidate(
		context.Context,
		usersdomain.InactivityWarningEligibility,
	) (usersdomain.InactivityWarningCandidate, bool, error)
	RecordUserInactivityWarning(
		context.Context,
		usersdomain.InactivityWarningReceipt,
	) (bool, error)
}

// ProcessUserInactivityWarning sends account-lifecycle warnings to active,
// non-system users after eight calendar months without a login.
func ProcessUserInactivityWarning(
	ctx context.Context,
	store UserInactivityWarningStore,
	log *logger.Logger,
	mailerService mailer.Service,
) error {
	return processUserInactivityWarningAt(ctx, store, log, mailerService, time.Now().UTC())
}

func processUserInactivityWarningAt(
	ctx context.Context,
	store UserInactivityWarningStore,
	log *logger.Logger,
	mailerService mailer.Service,
	now time.Time,
) error {
	if ctx == nil {
		return errors.New("user inactivity warning context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.ProcessUserInactivityWarning")
	defer span.End()
	if store == nil {
		return errors.New("user inactivity warning store is required")
	}
	if log == nil {
		return errors.New("user inactivity warning logger is required")
	}
	if mailerService == nil {
		return errors.New("user inactivity warning mailer is required")
	}
	if now.IsZero() {
		return errors.New("user inactivity warning clock is required")
	}

	now = now.UTC()
	inactiveBefore := now.AddDate(0, -8, 0)
	startedAt := time.Now()
	var cursor usersdomain.InactivityWarningCursor
	totalProcessed := 0
	totalEmailsSent := 0
	batchCount := 0

	log.Info(ctx, "Processing user inactivity warnings for users inactive for 8+ months")
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("process user inactivity warnings after %d candidates: %w", totalProcessed, err)
		}
		if batchCount >= userInactivityWarningMaxBatches {
			return fmt.Errorf(
				"process user inactivity warnings after %d candidates: %w",
				totalProcessed,
				errUserInactivityWarningBacklogRemaining,
			)
		}

		candidates, err := store.ListUserInactivityWarningCandidates(
			ctx,
			usersdomain.InactivityWarningQuery{
				InactiveBefore: inactiveBefore,
				Cursor:         cursor,
				BatchSize:      userInactivityWarningBatchSize,
			},
		)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("list user inactivity warning candidates: %w", err)
		}
		if len(candidates) == 0 {
			break
		}
		if len(candidates) > userInactivityWarningBatchSize {
			return fmt.Errorf(
				"list user inactivity warning candidates: got %d rows, want at most %d",
				len(candidates),
				userInactivityWarningBatchSize,
			)
		}

		batchCount++
		for _, candidate := range candidates {
			sent, sendErr := processUserInactivityWarningCandidate(
				ctx,
				store,
				mailerService,
				candidate,
				inactiveBefore,
				now,
			)
			if sendErr != nil {
				log.Error(
					ctx,
					"Failed to send user inactivity warning",
					"error", sendErr,
					"user_id", candidate.UserID,
				)
				continue
			}
			if sent {
				totalEmailsSent++
			}
		}
		totalProcessed += len(candidates)

		lastCandidate := candidates[len(candidates)-1]
		cursor = usersdomain.InactivityWarningCursor{
			LastLoginAt: lastCandidate.LastLoginAt,
			UserID:      lastCandidate.UserID,
			Valid:       true,
		}
		span.AddEvent("user inactivity warning batch processed", trace.WithAttributes(
			attribute.Int("batch", batchCount),
			attribute.Int("users.processed", len(candidates)),
		))
		if len(candidates) < userInactivityWarningBatchSize {
			break
		}
		if err := waitForUserInactivityWarningBatch(ctx); err != nil {
			return fmt.Errorf("pace user inactivity warning batches: %w", err)
		}
	}

	duration := time.Since(startedAt)
	span.AddEvent("user inactivity warnings completed", trace.WithAttributes(
		attribute.Int("users.processed", totalProcessed),
		attribute.Int("emails.sent", totalEmailsSent),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))
	log.Info(
		ctx,
		"User inactivity warning job completed",
		"emails_sent", totalEmailsSent,
		"users_processed", totalProcessed,
		"batches_processed", batchCount,
		"duration", duration,
	)
	return nil
}

func processUserInactivityWarningCandidate(
	ctx context.Context,
	store UserInactivityWarningStore,
	mailerService mailer.Service,
	candidate usersdomain.InactivityWarningCandidate,
	inactiveBefore time.Time,
	warningSentAt time.Time,
) (bool, error) {
	eligibleCandidate, eligible, err := store.GetEligibleUserInactivityWarningCandidate(
		ctx,
		usersdomain.InactivityWarningEligibility{
			UserID:         candidate.UserID,
			InactiveBefore: inactiveBefore,
		},
	)
	if err != nil {
		return false, fmt.Errorf("recheck user inactivity warning eligibility: %w", err)
	}
	if !eligible {
		return false, nil
	}

	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{eligibleCandidate.Email},
		Template: "users/inactivity_warning",
		Subject:  "Your account is scheduled for deactivation",
		Data: map[string]any{
			"UserName": eligibleCandidate.FullName,
			"LoginURL": "https://fortyone.app/login",
		},
	}); err != nil {
		return false, fmt.Errorf("send user inactivity warning email: %w", err)
	}

	recorded, err := store.RecordUserInactivityWarning(
		ctx,
		usersdomain.InactivityWarningReceipt{
			UserID:         eligibleCandidate.UserID,
			InactiveBefore: inactiveBefore,
			WarningSentAt:  warningSentAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("record user inactivity warning: %w", err)
	}
	if !recorded {
		return false, errors.New("record user inactivity warning: account eligibility changed after delivery")
	}
	return true, nil
}

func waitForUserInactivityWarningBatch(ctx context.Context) error {
	timer := time.NewTimer(userInactivityWarningBatchDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
