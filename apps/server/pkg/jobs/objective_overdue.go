package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	objectiveLeadBatchSize    = 100
	overdueGuidanceBatchDelay = 100 * time.Millisecond
)

// ObjectiveOverdueStore is the worker-owned persistence capability needed to
// page eligible recipients and load their current objective deadline signals.
type ObjectiveOverdueStore interface {
	ListOverdueObjectiveGuidanceRecipients(context.Context, time.Time, *objectivesdomain.OverdueGuidanceCursor, int) ([]objectivesdomain.OverdueGuidanceRecipient, error)
	ListOverdueObjectiveGuidanceItems(context.Context, time.Time, uuid.UUID, uuid.UUID) ([]objectivesdomain.OverdueGuidanceObjective, error)
}

// ProcessObjectiveOverdue processes overdue objectives and sends emails directly.
func ProcessObjectiveOverdue(ctx context.Context, store ObjectiveOverdueStore, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer) error {
	return processObjectiveOverdueAt(ctx, store, log, mailerService, copyGenerator, threader, time.Now().UTC())
}

func processObjectiveOverdueAt(ctx context.Context, store ObjectiveOverdueStore, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer, asOf time.Time) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessObjectiveOverdue")
	defer span.End()
	if store == nil {
		return errors.New("objective overdue store is required")
	}
	if log == nil {
		return errors.New("objective overdue logger is required")
	}
	if mailerService == nil {
		return errors.New("objective overdue mailer is required")
	}
	asOf, err := overdueGuidanceUTCDate(asOf)
	if err != nil {
		return err
	}

	log.Info(ctx, "Processing objective overdue notifications")
	startTime := time.Now()

	totalProcessed := 0
	totalEmailsCreated := 0
	batchCount := 0
	var cursor *objectivesdomain.OverdueGuidanceCursor

	for {
		nextBatch := batchCount + 1
		log.Info(ctx, fmt.Sprintf("Processing objective lead batch %d", nextBatch))

		leads, err := store.ListOverdueObjectiveGuidanceRecipients(ctx, asOf, cursor, objectiveLeadBatchSize)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get leads batch %d: %w", nextBatch, err)
		}
		if len(leads) == 0 {
			break
		}
		batchCount = nextBatch

		results, batchErr := processGuidanceEmailBatch(ctx, leads, func(batchCtx context.Context, lead objectivesdomain.OverdueGuidanceRecipient) guidanceEmailBatchResult {
			return processGuidanceEmailRecipient(batchCtx, func(attemptCtx context.Context) guidanceEmailBatchResult {
				objectives, objectivesErr := store.ListOverdueObjectiveGuidanceItems(attemptCtx, asOf, lead.LeadUserID, lead.WorkspaceID)
				if objectivesErr != nil {
					log.Error(attemptCtx, "Failed to get objectives for lead", "lead_id", lead.LeadUserID, "workspace_id", lead.WorkspaceID, "error", objectivesErr)
					return guidanceEmailBatchResult{Err: objectivesErr, Retryable: true}
				}
				if len(objectives) == 0 {
					return guidanceEmailBatchResult{Processed: true}
				}
				if sendErr := sendObjectiveOverdueEmailForLead(attemptCtx, log, mailerService, copyGenerator, threader, objectives); sendErr != nil {
					log.Error(attemptCtx, "Failed to send email", "lead_id", lead.LeadUserID, "error", sendErr)
					return guidanceEmailBatchResult{Err: sendErr}
				}
				return guidanceEmailBatchResult{Processed: true, Sent: true}
			})
		})
		if batchErr != nil {
			span.RecordError(batchErr)
			return fmt.Errorf("objective guidance batch cancelled: %w", batchErr)
		}
		for _, result := range results {
			if result.Processed {
				totalProcessed++
			}
			if result.Sent {
				totalEmailsCreated++
			}
		}
		if failureCount := guidanceEmailBatchFailureCount(results); failureCount > 0 {
			log.Error(ctx, "Objective recipients failed after in-job processing; continuing without retrying successful deliveries", "failed_recipients", failureCount, "batch", batchCount)
			span.AddEvent("objective recipient deliveries failed", trace.WithAttributes(attribute.Int("failed_recipients", failureCount)))
		}

		log.Info(ctx, fmt.Sprintf("Lead batch %d completed: %d leads processed", batchCount, len(leads)))
		lastLead := leads[len(leads)-1]
		cursor = &objectivesdomain.OverdueGuidanceCursor{
			LeadUserID:  lastLead.LeadUserID,
			WorkspaceID: lastLead.WorkspaceID,
		}
		if len(leads) < objectiveLeadBatchSize {
			break
		}
		if err := waitForNextOverdueGuidanceBatch(ctx); err != nil {
			span.RecordError(err)
			return fmt.Errorf("wait before objective lead batch %d: %w", batchCount+1, err)
		}
	}

	duration := time.Since(startTime)
	span.AddEvent("objective overdue job completed", trace.WithAttributes(
		attribute.Int64("leads.processed", int64(totalProcessed)),
		attribute.Int64("emails.created", int64(totalEmailsCreated)),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))
	log.Info(ctx, fmt.Sprintf("Objective overdue job completed: %d leads processed, %d emails created in %d batches over %v",
		totalProcessed, totalEmailsCreated, batchCount, duration))
	return nil
}

func overdueGuidanceUTCDate(value time.Time) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, errors.New("overdue guidance as-of time is required")
	}
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC), nil
}

func waitForNextOverdueGuidanceBatch(ctx context.Context) error {
	timer := time.NewTimer(overdueGuidanceBatchDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type OverdueObjective = objectivesdomain.OverdueGuidanceObjective
type OverdueKeyResult = objectivesdomain.OverdueGuidanceKeyResult
