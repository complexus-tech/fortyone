package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// VerificationTokenPurger is the worker-owned persistence capability for
// bounded verification-token retention.
type VerificationTokenPurger interface {
	PurgeExpiredVerificationTokens(context.Context, time.Time, int) (int64, error)
}

// PurgeExpiredTokens permanently deletes verification tokens whose expiry is
// more than seven days old, together with expired feedback contributor
// security artifacts.
func PurgeExpiredTokens(
	ctx context.Context,
	verificationTokens VerificationTokenPurger,
	feedbackStore feedback.MaintenanceStore,
	log *logger.Logger,
) error {
	return purgeExpiredTokensAt(ctx, verificationTokens, feedbackStore, log, time.Now().UTC())
}

func purgeExpiredTokensAt(
	ctx context.Context,
	verificationTokens VerificationTokenPurger,
	feedbackStore feedback.MaintenanceStore,
	log *logger.Logger,
	now time.Time,
) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeExpiredTokens")
	defer span.End()
	if verificationTokens == nil {
		return errors.New("verification token maintenance store is required")
	}
	if feedbackStore == nil {
		return errors.New("feedback maintenance store is required")
	}
	if log == nil {
		return errors.New("token maintenance logger is required")
	}
	if now.IsZero() {
		return errors.New("token maintenance clock is required")
	}
	now = now.UTC()
	retainedBefore := now.Add(-7 * 24 * time.Hour)

	log.Info(ctx, "Purging expired verification tokens and feedback security artifacts")
	verificationTokensDeleted, err := drainMaintenanceBatches(ctx, "purge expired verification tokens", func(ctx context.Context, batchSize int) (int64, error) {
		return verificationTokens.PurgeExpiredVerificationTokens(ctx, retainedBefore, batchSize)
	})
	if err != nil {
		span.RecordError(err)
		return err
	}

	feedbackResult, err := feedbackStore.PurgeExpiredContributorArtifacts(ctx, feedback.CoreContributorArtifactCutoffs{
		RetainedBefore: retainedBefore,
		ExpiredBefore:  now,
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("purge expired feedback contributor artifacts: %w", err)
	}

	span.AddEvent("tokens_deleted", trace.WithAttributes(
		attribute.Int64("verification_tokens", verificationTokensDeleted),
		attribute.Int64("feedback_artifacts", feedbackResult.TotalDeleted()),
	))
	log.Info(ctx, "Permanently deleted expired tokens",
		"verification_tokens", verificationTokensDeleted,
		"feedback_artifacts", feedbackResult.TotalDeleted(),
	)
	return nil
}
