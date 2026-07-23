package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	feedbackDigestBatchSize       = 100
	feedbackDigestDeliveryHour    = 9
	feedbackDigestConsistencyLag  = 5 * time.Minute
	feedbackDigestClaimStaleAfter = 2 * time.Hour
	feedbackDigestDateLayout      = "2006-01-02"
	feedbackDigestItemLimit       = 5
)

type feedbackDigestRecipient struct {
	UserID        uuid.UUID `db:"user_id"`
	UserEmail     string    `db:"user_email"`
	UserName      string    `db:"user_name"`
	Timezone      string    `db:"timezone"`
	WorkspaceID   uuid.UUID `db:"workspace_id"`
	WorkspaceName string    `db:"workspace_name"`
	WorkspaceSlug string    `db:"workspace_slug"`
}

type feedbackDigestSubscription struct {
	BoardID            uuid.UUID  `db:"board_id"`
	TeamID             uuid.UUID  `db:"team_id"`
	EmailFrequency     string     `db:"email_frequency"`
	CreatedAt          time.Time  `db:"created_at"`
	LastDigestSentAt   *time.Time `db:"last_digest_sent_at"`
	LastDigestCursorAt *time.Time `db:"last_digest_cursor_at"`
}

type feedbackDigestBoardWindow struct {
	feedbackDigestSubscription
	WindowStart time.Time
}

type feedbackDigestItem struct {
	ID                 uuid.UUID `db:"id"`
	TeamID             uuid.UUID `db:"team_id"`
	Title              string    `db:"title"`
	AuthorName         string    `db:"author_name"`
	TeamName           string    `db:"team_name"`
	TotalCount         int       `db:"total_count"`
	PendingReviewCount int       `db:"pending_review_count"`
}

// ProcessFeedbackDigestEmail sends due feedback digests at 09:00 in each
// recipient's timezone. A single delivery combines all due boards in a
// workspace, while each board keeps its own delivery cursor.
func ProcessFeedbackDigestEmail(ctx context.Context, db *sqlx.DB, log *logger.Logger, mailerService mailer.Service) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessFeedbackDigestEmail")
	defer span.End()

	now := time.Now().UTC()
	afterWorkspaceID := uuid.Nil
	afterUserID := uuid.Nil
	hasCursor := false
	var processingErrors []error

	for {
		recipients, err := getFeedbackDigestRecipients(ctx, db, feedbackDigestBatchSize, hasCursor, afterWorkspaceID, afterUserID)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get feedback digest recipients: %w", err)
		}
		if len(recipients) == 0 {
			break
		}

		for _, recipient := range recipients {
			if err := processFeedbackDigestRecipient(ctx, db, log, mailerService, recipient, now); err != nil {
				log.Error(ctx, "Failed to process feedback digest recipient",
					"recipient_id", recipient.UserID,
					"workspace_id", recipient.WorkspaceID,
					"error", err)
				processingErrors = append(processingErrors, err)
			}
		}

		lastRecipient := recipients[len(recipients)-1]
		afterWorkspaceID = lastRecipient.WorkspaceID
		afterUserID = lastRecipient.UserID
		hasCursor = true
	}

	if len(processingErrors) > 0 {
		err := errors.Join(processingErrors...)
		span.RecordError(err)
		return fmt.Errorf("one or more feedback digests failed: %w", err)
	}

	return nil
}

func processFeedbackDigestRecipient(
	ctx context.Context,
	db *sqlx.DB,
	log *logger.Logger,
	mailerService mailer.Service,
	recipient feedbackDigestRecipient,
	now time.Time,
) error {
	subscriptions, err := getFeedbackDigestSubscriptions(ctx, db, recipient.UserID, recipient.WorkspaceID)
	if err != nil {
		return fmt.Errorf("get subscriptions for recipient %s: %w", recipient.UserID, err)
	}

	location := feedbackDigestLocation(recipient.Timezone)
	dueSubscriptions := dueFeedbackDigestSubscriptions(now, location, subscriptions)
	if len(dueSubscriptions) == 0 {
		return nil
	}

	localDate := now.In(location).Format(feedbackDigestDateLayout)
	windowEnd := now.Add(-feedbackDigestConsistencyLag)
	windowStart := earliestFeedbackDigestWindowStart(dueSubscriptions, windowEnd)
	deliveryID, claimed, err := claimFeedbackDigestDelivery(
		ctx,
		db,
		recipient.WorkspaceID,
		recipient.UserID,
		localDate,
		windowStart,
		windowEnd,
	)
	if err != nil {
		return fmt.Errorf("claim delivery for recipient %s: %w", recipient.UserID, err)
	}
	if !claimed {
		return nil
	}

	items, err := getFeedbackDigestItems(ctx, db, recipient, dueSubscriptions, windowEnd)
	if err != nil {
		return failFeedbackDigestDelivery(ctx, db, deliveryID, fmt.Errorf("get digest items: %w", err))
	}

	boardIDs := feedbackDigestBoardIDs(dueSubscriptions)
	if len(items) == 0 {
		if err := completeFeedbackDigestDelivery(ctx, db, deliveryID, recipient, boardIDs, now, windowEnd, "skipped", 0); err != nil {
			return failFeedbackDigestDelivery(ctx, db, deliveryID, fmt.Errorf("complete empty digest: %w", err))
		}
		log.Info(ctx, "Skipped empty feedback digest",
			"recipient_id", recipient.UserID,
			"workspace_id", recipient.WorkspaceID,
			"local_date", localDate)
		return nil
	}

	itemCount := items[0].TotalCount
	if err := sendFeedbackDigestEmail(ctx, mailerService, deliveryID, recipient, items); err != nil {
		return failFeedbackDigestDelivery(ctx, db, deliveryID, err)
	}

	if err := completeFeedbackDigestDelivery(ctx, db, deliveryID, recipient, boardIDs, now, windowEnd, "sent", itemCount); err != nil {
		return failFeedbackDigestDelivery(ctx, db, deliveryID, fmt.Errorf("complete sent digest: %w", err))
	}

	log.Info(ctx, "Successfully sent feedback digest",
		"recipient_id", recipient.UserID,
		"workspace_id", recipient.WorkspaceID,
		"local_date", localDate,
		"item_count", itemCount)
	return nil
}

func getFeedbackDigestRecipients(
	ctx context.Context,
	db *sqlx.DB,
	batchSize int,
	hasCursor bool,
	afterWorkspaceID uuid.UUID,
	afterUserID uuid.UUID,
) ([]feedbackDigestRecipient, error) {
	params := map[string]any{
		"batch_size":         batchSize,
		"has_cursor":         hasCursor,
		"after_workspace_id": afterWorkspaceID,
		"after_user_id":      afterUserID,
	}

	stmt, err := db.PrepareNamedContext(ctx, feedbackDigestRecipientsQuery())
	if err != nil {
		return nil, fmt.Errorf("prepare feedback digest recipients query: %w", err)
	}
	defer stmt.Close()

	var recipients []feedbackDigestRecipient
	if err := stmt.SelectContext(ctx, &recipients, params); err != nil {
		return nil, fmt.Errorf("execute feedback digest recipients query: %w", err)
	}
	return recipients, nil
}

func feedbackDigestRecipientsQuery() string {
	return `
		SELECT DISTINCT
			u.user_id,
			u.email AS user_email,
			COALESCE(NULLIF(TRIM(u.full_name), ''), NULLIF(TRIM(u.username), ''), u.email) AS user_name,
			COALESCE(NULLIF(TRIM(u.timezone), ''), 'UTC') AS timezone,
			w.workspace_id,
			w.name AS workspace_name,
			w.slug AS workspace_slug
		FROM feedback_board_subscriptions fbs
		INNER JOIN feedback_boards fb ON fb.id = fbs.board_id
		INNER JOIN teams t
			ON t.team_id = fb.team_id
			AND t.workspace_id = fb.workspace_id
		INNER JOIN team_members tm
			ON tm.team_id = fb.team_id
			AND tm.user_id = fbs.user_id
		INNER JOIN workspace_members wm
			ON wm.workspace_id = fb.workspace_id
			AND wm.user_id = fbs.user_id
			AND wm.role IN ('admin', 'member')
		INNER JOIN users u
			ON u.user_id = fbs.user_id
			AND u.is_active = true
			AND u.is_system = false
		INNER JOIN workspaces w
			ON w.workspace_id = fb.workspace_id
			AND w.deleted_at IS NULL
		WHERE fbs.email_frequency IN ('daily', 'weekly')
			AND NULLIF(TRIM(u.email), '') IS NOT NULL
			AND (
				:has_cursor = false
				OR w.workspace_id > :after_workspace_id
				OR (w.workspace_id = :after_workspace_id AND u.user_id > :after_user_id)
			)
		ORDER BY w.workspace_id, u.user_id
		LIMIT :batch_size;
	`
}

func getFeedbackDigestSubscriptions(ctx context.Context, db *sqlx.DB, recipientID, workspaceID uuid.UUID) ([]feedbackDigestSubscription, error) {
	var subscriptions []feedbackDigestSubscription
	if err := db.SelectContext(ctx, &subscriptions, feedbackDigestSubscriptionsQuery(), recipientID, workspaceID); err != nil {
		return nil, fmt.Errorf("execute feedback digest subscriptions query: %w", err)
	}
	return subscriptions, nil
}

func feedbackDigestSubscriptionsQuery() string {
	return `
		SELECT
			fbs.board_id,
			fb.team_id,
			fbs.email_frequency,
			fbs.created_at,
			fbs.last_digest_sent_at,
			fbs.last_digest_cursor_at
		FROM feedback_board_subscriptions fbs
		INNER JOIN feedback_boards fb ON fb.id = fbs.board_id
		INNER JOIN teams t
			ON t.team_id = fb.team_id
			AND t.workspace_id = fb.workspace_id
		INNER JOIN team_members tm
			ON tm.team_id = fb.team_id
			AND tm.user_id = fbs.user_id
		INNER JOIN workspace_members wm
			ON wm.workspace_id = fb.workspace_id
			AND wm.user_id = fbs.user_id
			AND wm.role IN ('admin', 'member')
		WHERE fbs.user_id = $1
			AND fb.workspace_id = $2
			AND fbs.email_frequency IN ('daily', 'weekly')
		ORDER BY fbs.board_id;
	`
}

func feedbackDigestLocation(timezone string) *time.Location {
	location, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		return time.UTC
	}
	return location
}

func dueFeedbackDigestSubscriptions(now time.Time, location *time.Location, subscriptions []feedbackDigestSubscription) []feedbackDigestBoardWindow {
	localNow := now.In(location)
	due := make([]feedbackDigestBoardWindow, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		if !isFeedbackDigestSubscriptionDue(localNow, location, subscription) {
			continue
		}

		windowStart := subscription.CreatedAt.UTC()
		if subscription.LastDigestCursorAt != nil {
			windowStart = subscription.LastDigestCursorAt.UTC()
		}
		due = append(due, feedbackDigestBoardWindow{
			feedbackDigestSubscription: subscription,
			WindowStart:                windowStart,
		})
	}
	return due
}

func isFeedbackDigestSubscriptionDue(
	localNow time.Time,
	location *time.Location,
	subscription feedbackDigestSubscription,
) bool {
	if subscription.EmailFrequency != "daily" && subscription.EmailFrequency != "weekly" {
		return false
	}

	periodStart := time.Date(
		localNow.Year(),
		localNow.Month(),
		localNow.Day(),
		feedbackDigestDeliveryHour,
		0,
		0,
		0,
		location,
	)
	if subscription.EmailFrequency == "weekly" {
		daysSinceMonday := (int(localNow.Weekday()) - int(time.Monday) + 7) % 7
		periodStart = periodStart.AddDate(0, 0, -daysSinceMonday)
	}
	if localNow.Before(periodStart) {
		return false
	}
	if subscription.LastDigestSentAt != nil {
		return subscription.LastDigestSentAt.Before(periodStart.UTC())
	}

	// A reviewer who subscribes after this period's scheduled delivery starts
	// with the next period instead of receiving an immediate partial digest.
	return !subscription.CreatedAt.After(periodStart.UTC())
}

func earliestFeedbackDigestWindowStart(subscriptions []feedbackDigestBoardWindow, windowEnd time.Time) time.Time {
	earliest := subscriptions[0].WindowStart
	for _, subscription := range subscriptions[1:] {
		if subscription.WindowStart.Before(earliest) {
			earliest = subscription.WindowStart
		}
	}
	if !earliest.Before(windowEnd) {
		return windowEnd.Add(-time.Millisecond)
	}
	return earliest
}

func claimFeedbackDigestDelivery(
	ctx context.Context,
	db *sqlx.DB,
	workspaceID uuid.UUID,
	recipientID uuid.UUID,
	localDate string,
	windowStart time.Time,
	windowEnd time.Time,
) (uuid.UUID, bool, error) {
	var deliveryID uuid.UUID
	err := db.GetContext(ctx, &deliveryID, feedbackDigestClaimQuery(), workspaceID, recipientID, localDate, windowStart, windowEnd)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("execute feedback digest delivery claim: %w", err)
	}
	return deliveryID, true, nil
}

func feedbackDigestClaimQuery() string {
	return fmt.Sprintf(`
		INSERT INTO feedback_digest_deliveries (
			workspace_id,
			recipient_id,
			local_date,
			status,
			window_start,
			window_end
		)
		VALUES ($1, $2, $3, 'processing', $4, $5)
		ON CONFLICT (workspace_id, recipient_id, local_date)
		DO UPDATE SET
			status = 'processing',
			window_start = EXCLUDED.window_start,
			window_end = EXCLUDED.window_end,
			item_count = 0,
			sent_at = NULL,
			last_error = NULL,
			updated_at = NOW()
		WHERE feedback_digest_deliveries.status = 'failed'
			OR (
				feedback_digest_deliveries.status = 'processing'
				AND feedback_digest_deliveries.updated_at < NOW() - INTERVAL '%d seconds'
			)
		RETURNING id;
	`, int(feedbackDigestClaimStaleAfter.Seconds()))
}

func getFeedbackDigestItems(
	ctx context.Context,
	db *sqlx.DB,
	recipient feedbackDigestRecipient,
	subscriptions []feedbackDigestBoardWindow,
	windowEnd time.Time,
) ([]feedbackDigestItem, error) {
	boardIDs := make([]string, 0, len(subscriptions))
	windowStarts := make([]time.Time, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		boardIDs = append(boardIDs, subscription.BoardID.String())
		windowStarts = append(windowStarts, subscription.WindowStart)
	}

	var items []feedbackDigestItem
	if err := db.SelectContext(
		ctx,
		&items,
		feedbackDigestItemsQuery(),
		recipient.UserID,
		recipient.WorkspaceID,
		pq.Array(boardIDs),
		pq.Array(windowStarts),
		windowEnd,
		feedbackDigestItemLimit,
	); err != nil {
		return nil, fmt.Errorf("execute feedback digest items query: %w", err)
	}
	return items, nil
}

func feedbackDigestItemsQuery() string {
	return `
		WITH board_windows AS (
			SELECT board_id, window_start
			FROM UNNEST($3::uuid[], $4::timestamptz[]) AS windows(board_id, window_start)
		), eligible_items AS (
			SELECT
				fi.id,
				fb.team_id,
				fi.title,
				COALESCE(
					NULLIF(TRIM(author.full_name), ''),
					NULLIF(TRIM(author.username), ''),
					NULLIF(TRIM(author.email), ''),
					'Customer'
				) AS author_name,
				t.name AS team_name,
				fi.status,
				fi.created_at
			FROM board_windows bw
			INNER JOIN feedback_items fi ON fi.board_id = bw.board_id
			INNER JOIN feedback_boards fb
				ON fb.id = fi.board_id
				AND fb.workspace_id = fi.workspace_id
			INNER JOIN feedback_board_subscriptions fbs
				ON fbs.board_id = fb.id
				AND fbs.user_id = $1
			INNER JOIN teams t
				ON t.team_id = fb.team_id
				AND t.workspace_id = fb.workspace_id
			INNER JOIN team_members tm
				ON tm.team_id = fb.team_id
				AND tm.user_id = fbs.user_id
			INNER JOIN workspace_members wm
				ON wm.workspace_id = fb.workspace_id
				AND wm.user_id = fbs.user_id
				AND wm.role IN ('admin', 'member')
			LEFT JOIN users author ON author.user_id = fi.author_id
			WHERE fi.workspace_id = $2
				AND fi.deleted_at IS NULL
				AND fi.submission_source IN ('portal', 'widget', 'integration')
				AND fi.created_at > bw.window_start
				AND fi.created_at <= $5
		)
		SELECT
			id,
			team_id,
			title,
			author_name,
			team_name,
			COUNT(*) OVER ()::int AS total_count,
			COUNT(*) FILTER (WHERE status IN ('pending', 'reviewing')) OVER ()::int AS pending_review_count
		FROM eligible_items
		ORDER BY created_at DESC, id DESC
		LIMIT $6;
	`
}

func feedbackDigestBoardIDs(subscriptions []feedbackDigestBoardWindow) []string {
	boardIDs := make([]string, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		boardIDs = append(boardIDs, subscription.BoardID.String())
	}
	return boardIDs
}

func completeFeedbackDigestDelivery(
	ctx context.Context,
	db *sqlx.DB,
	deliveryID uuid.UUID,
	recipient feedbackDigestRecipient,
	boardIDs []string,
	deliveryAt time.Time,
	windowEnd time.Time,
	status string,
	itemCount int,
) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin feedback digest completion transaction: %w", err)
	}
	defer tx.Rollback()

	cursorResult, err := tx.ExecContext(ctx, `
		UPDATE feedback_board_subscriptions fbs
		SET last_digest_sent_at = $1,
			last_digest_cursor_at = $2,
			updated_at = NOW()
		FROM feedback_boards fb
		INNER JOIN teams t
			ON t.team_id = fb.team_id
			AND t.workspace_id = fb.workspace_id
		INNER JOIN team_members tm
			ON tm.team_id = fb.team_id
			AND tm.user_id = $3
		INNER JOIN workspace_members wm
			ON wm.workspace_id = fb.workspace_id
			AND wm.user_id = $3
			AND wm.role IN ('admin', 'member')
		WHERE fbs.board_id = fb.id
			AND fbs.user_id = $3
			AND fb.workspace_id = $4
			AND fbs.board_id = ANY($5::uuid[]);
	`, deliveryAt, windowEnd, recipient.UserID, recipient.WorkspaceID, pq.Array(boardIDs))
	if err != nil {
		return fmt.Errorf("advance feedback digest subscription cursors: %w", err)
	}
	if _, err := cursorResult.RowsAffected(); err != nil {
		return fmt.Errorf("read feedback digest cursor update result: %w", err)
	}

	sentAt := sql.NullTime{}
	if status == "sent" {
		sentAt = sql.NullTime{Time: deliveryAt, Valid: true}
	}
	deliveryResult, err := tx.ExecContext(ctx, `
		UPDATE feedback_digest_deliveries
		SET status = $1,
			item_count = $2,
			sent_at = $3,
			last_error = NULL,
			updated_at = NOW()
		WHERE id = $4
			AND status = 'processing';
	`, status, itemCount, sentAt, deliveryID)
	if err != nil {
		return fmt.Errorf("update feedback digest delivery: %w", err)
	}
	rowsAffected, err := deliveryResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("read feedback digest delivery update result: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("feedback digest delivery %s is no longer processing", deliveryID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit feedback digest completion: %w", err)
	}
	return nil
}

func failFeedbackDigestDelivery(ctx context.Context, db *sqlx.DB, deliveryID uuid.UUID, processingErr error) error {
	_, updateErr := db.ExecContext(ctx, `
		UPDATE feedback_digest_deliveries
		SET status = 'failed',
			last_error = LEFT($1::text, 2000),
			updated_at = NOW()
		WHERE id = $2
			AND status = 'processing';
	`, processingErr.Error(), deliveryID)
	if updateErr != nil {
		return errors.Join(processingErr, fmt.Errorf("mark feedback digest delivery failed: %w", updateErr))
	}
	return processingErr
}

func sendFeedbackDigestEmail(
	ctx context.Context,
	mailerService mailer.Service,
	deliveryID uuid.UUID,
	recipient feedbackDigestRecipient,
	items []feedbackDigestItem,
) error {
	if len(items) == 0 {
		return nil
	}

	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", recipient.WorkspaceSlug)
	subject := feedbackDigestSubject(items[0].TotalCount, recipient.WorkspaceName)
	data := map[string]any{
		"UserName":                 recipient.UserName,
		"UserEmail":                recipient.UserEmail,
		"WorkspaceName":            recipient.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        subject,
		"NotificationMessage":      formatFeedbackDigestEmailContent(items, workspaceURL),
		"NotificationType":         "feedback_digest",
		"NotificationCTAURL":       fmt.Sprintf("%s/teams/%s/feedback", workspaceURL, items[0].TeamID),
		"NotificationCTALabel":     "Review feedback",
		"NotificationsSettingsURL": fmt.Sprintf("%s/settings/workspace/feedback", workspaceURL),
	}

	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:        []string{recipient.UserEmail},
		Template:  "notifications/notification",
		Subject:   subject,
		Data:      data,
		MessageID: feedbackDigestMessageID(deliveryID),
	}); err != nil {
		return fmt.Errorf("send feedback digest email: %w", err)
	}
	return nil
}

func feedbackDigestMessageID(deliveryID uuid.UUID) string {
	return fmt.Sprintf("<feedback-digest-%s@fortyone.app>", deliveryID)
}

func feedbackDigestSubject(itemCount int, workspaceName string) string {
	return fmt.Sprintf(
		"%d new feedback %s in %s",
		itemCount,
		pluralize(itemCount, "item", "items"),
		workspaceName,
	)
}

func formatFeedbackDigestEmailContent(items []feedbackDigestItem, workspaceURL string) string {
	if len(items) == 0 {
		return ""
	}

	totalCount := items[0].TotalCount
	pendingReviewCount := items[0].PendingReviewCount
	intro := feedbackDigestIntro(totalCount, pendingReviewCount)
	displayedCount := min(len(items), feedbackDigestItemLimit)
	rows := make([]string, 0, displayedCount+1)
	for _, item := range items[:displayedCount] {
		itemURL := fmt.Sprintf("%s/teams/%s/feedback/%s", workspaceURL, item.TeamID, item.ID)
		rows = append(rows, fmt.Sprintf(
			"%s by %s in %s",
			formatEmailLink(itemURL, item.Title),
			html.EscapeString(item.AuthorName),
			html.EscapeString(item.TeamName),
		))
	}
	if remainingCount := totalCount - displayedCount; remainingCount > 0 {
		rows = append(rows, fmt.Sprintf(
			"%s more %s waiting in Feedback",
			formatEmailStrong(fmt.Sprintf("+%d", remainingCount)),
			pluralize(remainingCount, "item is", "items are"),
		))
	}
	return formatCompactNotificationRows(intro, rows)
}

func feedbackDigestIntro(itemCount, pendingReviewCount int) string {
	itemVerb := "items were"
	if itemCount == 1 {
		itemVerb = "item was"
	}
	intro := fmt.Sprintf("%d new %s submitted.", itemCount, itemVerb)
	if pendingReviewCount == 0 {
		return intro + " Everything is already being handled."
	}
	reviewVerb := "need"
	if pendingReviewCount == 1 {
		reviewVerb = "needs"
	}
	return fmt.Sprintf("%s %d still %s review.", intro, pendingReviewCount, reviewVerb)
}
