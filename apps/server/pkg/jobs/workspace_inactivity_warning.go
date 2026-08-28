package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	workspaceInactivityWarningBatchSize  = 100
	workspaceInactivityWarningBatchDelay = 100 * time.Millisecond
)

// WorkspaceInactivityWarningStore is the worker-owned persistence capability
// for keyset-paged inactivity warnings.
type WorkspaceInactivityWarningStore interface {
	ListWorkspaceInactivityWarningCandidates(
		context.Context,
		workspacedomain.InactivityWarningQuery,
	) ([]workspacedomain.InactivityWarningCandidate, error)
	RecordWorkspaceInactivityWarning(context.Context, workspacedomain.InactivityWarningReceipt) error
}

// ProcessWorkspaceInactivityWarning sends one warning to every active admin of
// a workspace after six calendar months without access.
func ProcessWorkspaceInactivityWarning(
	ctx context.Context,
	store WorkspaceInactivityWarningStore,
	log *logger.Logger,
	mailerService mailer.Service,
) error {
	return processWorkspaceInactivityWarningAt(ctx, store, log, mailerService, time.Now().UTC())
}

func processWorkspaceInactivityWarningAt(
	ctx context.Context,
	store WorkspaceInactivityWarningStore,
	log *logger.Logger,
	mailerService mailer.Service,
	now time.Time,
) error {
	if ctx == nil {
		return errors.New("workspace inactivity warning context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.ProcessWorkspaceInactivityWarning")
	defer span.End()
	if store == nil {
		return errors.New("workspace inactivity warning store is required")
	}
	if log == nil {
		return errors.New("workspace inactivity warning logger is required")
	}
	if mailerService == nil {
		return errors.New("workspace inactivity warning mailer is required")
	}
	if now.IsZero() {
		return errors.New("workspace inactivity warning clock is required")
	}

	now = now.UTC()
	inactiveBefore := now.AddDate(0, -6, 0)
	startedAt := time.Now()
	var cursor workspacedomain.InactivityCursor
	totalProcessed := 0
	totalEmailsSent := 0
	batchCount := 0

	log.Info(ctx, "Processing workspace inactivity warnings for workspaces inactive for 6+ months")
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("process workspace inactivity warnings after %d candidates: %w", totalProcessed, err)
		}

		candidates, err := store.ListWorkspaceInactivityWarningCandidates(
			ctx,
			workspacedomain.InactivityWarningQuery{
				InactiveBefore: inactiveBefore,
				Cursor:         cursor,
				BatchSize:      workspaceInactivityWarningBatchSize,
			},
		)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("list workspace inactivity warning candidates: %w", err)
		}
		if len(candidates) == 0 {
			break
		}
		if len(candidates) > workspaceInactivityWarningBatchSize {
			return fmt.Errorf(
				"list workspace inactivity warning candidates: got %d rows, want at most %d",
				len(candidates),
				workspaceInactivityWarningBatchSize,
			)
		}

		batchCount++
		for _, candidate := range candidates {
			if err := sendWorkspaceInactivityWarning(
				ctx,
				store,
				mailerService,
				candidate,
				inactiveBefore,
				now,
			); err != nil {
				log.Error(
					ctx,
					"Failed to send workspace inactivity warning",
					"error", err,
					"workspace_id", candidate.WorkspaceID,
					"workspace_name", candidate.Name,
				)
				continue
			}
			totalEmailsSent++
		}
		totalProcessed += len(candidates)

		lastCandidate := candidates[len(candidates)-1]
		cursor = workspacedomain.InactivityCursor{
			LastAccessedAt: lastCandidate.LastAccessedAt,
			WorkspaceID:    lastCandidate.WorkspaceID,
			Valid:          true,
		}
		span.AddEvent("workspace inactivity warning batch processed", trace.WithAttributes(
			attribute.Int("batch", batchCount),
			attribute.Int("workspaces.processed", len(candidates)),
		))
		if len(candidates) < workspaceInactivityWarningBatchSize {
			break
		}
		if err := waitForWorkspaceWarningBatch(ctx); err != nil {
			return fmt.Errorf("pace workspace inactivity warning batches: %w", err)
		}
	}

	duration := time.Since(startedAt)
	span.AddEvent("workspace inactivity warnings completed", trace.WithAttributes(
		attribute.Int("workspaces.processed", totalProcessed),
		attribute.Int("emails.sent", totalEmailsSent),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))
	log.Info(
		ctx,
		"Workspace inactivity warning job completed",
		"emails_sent", totalEmailsSent,
		"workspaces_processed", totalProcessed,
		"batches_processed", batchCount,
		"duration", duration,
	)
	return nil
}

func sendWorkspaceInactivityWarning(
	ctx context.Context,
	store WorkspaceInactivityWarningStore,
	mailerService mailer.Service,
	candidate workspacedomain.InactivityWarningCandidate,
	inactiveBefore time.Time,
	warningSentAt time.Time,
) error {
	if len(candidate.AdminEmails) == 0 {
		return fmt.Errorf("no admin emails found for workspace %s", candidate.Name)
	}

	data := map[string]any{
		"WorkspaceName": candidate.Name,
		"WorkspaceURL":  fmt.Sprintf("https://%s.fortyone.app", candidate.Slug),
	}
	subject := fmt.Sprintf("%s workspace scheduled for deletion", candidate.Name)
	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       append([]string(nil), candidate.AdminEmails...),
		Template: "workspaces/inactivity_warning",
		Subject:  subject,
		Data:     data,
	}); err != nil {
		return fmt.Errorf("send workspace inactivity warning email: %w", err)
	}

	if err := store.RecordWorkspaceInactivityWarning(
		ctx,
		workspacedomain.InactivityWarningReceipt{
			WorkspaceID:    candidate.WorkspaceID,
			InactiveBefore: inactiveBefore,
			WarningSentAt:  warningSentAt,
		},
	); err != nil {
		return fmt.Errorf("record workspace inactivity warning: %w", err)
	}
	return nil
}

func waitForWorkspaceWarningBatch(ctx context.Context) error {
	timer := time.NewTimer(workspaceInactivityWarningBatchDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
