package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessUserInactivityWarning sends warning emails to users inactive for 8+ months
func ProcessUserInactivityWarning(ctx context.Context, db *sqlx.DB, log *logger.Logger, brevoService *brevo.Service) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessUserInactivityWarning")
	defer span.End()

	log.Info(ctx, "Processing user inactivity warnings for users inactive for 8+ months")
	startTime := time.Now()

	const userBatchSize = 100 // Process 100 users at a time
	totalProcessed := 0
	totalEmailsSent := 0
	batchCount := 0

	for {
		batchCount++
		log.Info(ctx, fmt.Sprintf("Processing user batch %d", batchCount))

		// Get next batch of inactive users
		users, err := getInactiveUsersBatch(ctx, db, userBatchSize, batchCount-1)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get users batch %d: %w", batchCount, err)
		}

		if len(users) == 0 {
			break // No more users
		}

		// Process each user in this batch
		for _, user := range users {
			if err := sendUserInactivityWarning(ctx, brevoService, user.Email, user.FullName, user.LastLoginAt); err != nil {
				log.Error(ctx, "Failed to send user inactivity warning", "error", err, "user_id", user.UserID, "email", user.Email)
				continue
			}
			totalEmailsSent++
		}
		totalProcessed += len(users)

		log.Info(ctx, fmt.Sprintf("User batch %d completed: %d users processed", batchCount, len(users)))

		// Small delay to avoid overwhelming the email service
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)

	span.AddEvent("user inactivity warnings completed", trace.WithAttributes(
		attribute.Int("users.processed", totalProcessed),
		attribute.Int("emails.sent", totalEmailsSent),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))

	log.Info(ctx, fmt.Sprintf("User inactivity warning job completed: %d emails sent to %d users in %d batches over %v", totalEmailsSent, totalProcessed, batchCount, duration))
	return nil
}

// InactiveUser represents a user who hasn't been active for 8+ months
type InactiveUser struct {
	UserID      string    `db:"user_id"`
	Email       string    `db:"email"`
	FullName    string    `db:"full_name"`
	LastLoginAt time.Time `db:"last_login_at"`
}

// getInactiveUsersBatch gets a batch of users who haven't been active for 8+ months
func getInactiveUsersBatch(ctx context.Context, db *sqlx.DB, batchSize int, offset int) ([]InactiveUser, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getInactiveUsersBatch")
	defer span.End()

	query := `
		SELECT 
			u.user_id,
			u.email,
			u.full_name,
			u.last_login_at
		FROM users u
		WHERE u.last_login_at < NOW() - INTERVAL '8 months'
		AND u.is_active = true
		ORDER BY u.last_login_at ASC
		LIMIT :batch_size OFFSET :offset
	`

	params := map[string]any{
		"batch_size": batchSize,
		"offset":     offset * batchSize,
	}

	var users []InactiveUser
	err := db.SelectContext(ctx, &users, query, params)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query inactive users batch: %w", err)
	}

	span.AddEvent("inactive users batch retrieved", trace.WithAttributes(
		attribute.Int("users.count", len(users)),
		attribute.Int("batch_size", batchSize),
		attribute.Int("offset", offset),
	))

	return users, nil
}

// sendUserInactivityWarning sends a warning email to inactive user
func sendUserInactivityWarning(ctx context.Context, brevoService *brevo.Service, email, fullName string, lastLoginAt time.Time) error {
	// Calculate days since last login
	daysSinceLogin := int(time.Since(lastLoginAt).Hours() / 24)

	// Email template parameters
	brevoParams := map[string]any{
		"USER_NAME":         fullName,
		"DAYS_INACTIVE":     daysSinceLogin,
		"DEACTIVATION_DATE": time.Now().AddDate(0, 0, 30).Format("January 2, 2006"),
	}

	// Send templated email via Brevo service
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateUserInactivityWarning, // You'll need to add this template ID
		To: []brevo.EmailRecipient{
			{
				Email: email,
				Name:  fullName,
			},
		},
		Params: brevoParams,
	}

	if err := brevoService.SendTemplatedEmail(ctx, req); err != nil {
		return fmt.Errorf("failed to send user inactivity warning email: %w", err)
	}

	return nil
}
