package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessWorkspaceInactivityWarning sends warning emails to workspace admins for workspaces inactive for 6+ months
func ProcessWorkspaceInactivityWarning(ctx context.Context, db *sqlx.DB, log *logger.Logger, brevoService *brevo.Service) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessWorkspaceInactivityWarning")
	defer span.End()

	log.Info(ctx, "Processing workspace inactivity warnings for workspaces inactive for 6+ months")
	startTime := time.Now()

	const workspaceBatchSize = 100 // Process 100 admin-workspace pairs at a time
	totalProcessed := 0
	totalEmailsSent := 0
	batchCount := 0

	for {
		batchCount++
		log.Info(ctx, fmt.Sprintf("Processing workspace batch %d", batchCount))

		// Get next batch of inactive workspaces with their admins
		workspaces, err := getInactiveWorkspacesBatch(ctx, db, workspaceBatchSize, batchCount-1)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get workspaces batch %d: %w", batchCount, err)
		}

		if len(workspaces) == 0 {
			break // No more workspaces
		}

		// Process each workspace in this batch
		for _, ws := range workspaces {
			if err := sendWorkspaceInactivityWarning(ctx, brevoService, ws); err != nil {
				log.Error(ctx, "Failed to send workspace inactivity warning", "error", err, "workspace_id", ws.WorkspaceID, "workspace_name", ws.Name)
				continue
			}
			totalEmailsSent++
		}
		totalProcessed += len(workspaces)

		log.Info(ctx, fmt.Sprintf("Workspace batch %d completed: %d workspaces processed", batchCount, len(workspaces)))

		// Small delay to avoid overwhelming the email service
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)

	span.AddEvent("workspace inactivity warnings completed", trace.WithAttributes(
		attribute.Int("workspaces.processed", totalProcessed),
		attribute.Int("emails.sent", totalEmailsSent),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))

	log.Info(ctx, fmt.Sprintf("Workspace inactivity warning job completed: %d emails sent to %d workspaces in %d batches over %v", totalEmailsSent, totalProcessed, batchCount, duration))
	return nil
}

// InactiveWorkspace represents a workspace that hasn't been active for 6+ months with its admin emails
type InactiveWorkspace struct {
	WorkspaceID    uuid.UUID       `db:"workspace_id"`
	Name           string          `db:"name"`
	Slug           string          `db:"slug"`
	LastAccessedAt time.Time       `db:"last_accessed_at"`
	AdminEmails    json.RawMessage `db:"admin_emails"`
}

// getInactiveWorkspacesBatch gets a batch of workspaces that haven't been active for 6+ months with their admin emails
func getInactiveWorkspacesBatch(ctx context.Context, db *sqlx.DB, batchSize int, offset int) ([]InactiveWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getInactiveWorkspacesBatch")
	defer span.End()

	query := `
		SELECT 
			w.workspace_id,
			w.name,
			w.slug,
			w.last_accessed_at,
			COALESCE(
				(
					SELECT json_agg(u.email)
					FROM workspace_members wm
					INNER JOIN users u ON wm.user_id = u.user_id
					WHERE wm.workspace_id = w.workspace_id
					AND wm.role = 'admin'
					AND u.is_active = true
				), '[]'
			) AS admin_emails
		FROM workspaces w
		WHERE w.last_accessed_at < NOW() - INTERVAL '6 months'
		AND w.deleted_at IS NULL
		ORDER BY w.last_accessed_at ASC
		LIMIT :batch_size OFFSET :offset
	`

	params := map[string]any{
		"batch_size": batchSize,
		"offset":     offset * batchSize,
	}

	var workspaces []InactiveWorkspace
	err := db.SelectContext(ctx, &workspaces, query, params)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query inactive workspaces batch: %w", err)
	}

	span.AddEvent("inactive workspaces batch retrieved", trace.WithAttributes(
		attribute.Int("workspaces.count", len(workspaces)),
		attribute.Int("batch_size", batchSize),
		attribute.Int("offset", offset),
	))

	return workspaces, nil
}

// sendWorkspaceInactivityWarning sends a warning email to all workspace admins
func sendWorkspaceInactivityWarning(ctx context.Context, brevoService *brevo.Service, ws InactiveWorkspace) error {
	// Parse admin emails from JSON
	var adminEmails []string
	if err := json.Unmarshal(ws.AdminEmails, &adminEmails); err != nil {
		return fmt.Errorf("failed to unmarshal admin emails: %w", err)
	}

	// Skip if no admin emails
	if len(adminEmails) == 0 {
		return fmt.Errorf("no admin emails found for workspace %s", ws.Name)
	}

	// Calculate days since last access
	daysSinceAccess := int(time.Since(ws.LastAccessedAt).Hours() / 24)

	// Email template parameters
	brevoParams := map[string]any{
		"WORKSPACE_NAME": ws.Name,
		"WORKSPACE_SLUG": ws.Slug,
		"DAYS_INACTIVE":  daysSinceAccess,
		"DELETION_DATE":  time.Now().AddDate(0, 0, 30).Format("January 2, 2006"),
	}

	// Create email recipients for all admins
	recipients := make([]brevo.EmailRecipient, len(adminEmails))
	for i, email := range adminEmails {
		recipients[i] = brevo.EmailRecipient{Email: email}
	}

	// Send templated email to all admins
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateWorkspaceInactivityWarning,
		To:         recipients,
		Params:     brevoParams,
	}

	if err := brevoService.SendTemplatedEmail(ctx, req); err != nil {
		return fmt.Errorf("failed to send workspace inactivity warning email: %w", err)
	}

	return nil
}
