package jobs

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"
	_ "time/tzdata"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	htmlparser "golang.org/x/net/html"
)

const (
	feedbackDigestBatchSize       = 100
	feedbackDigestDeliveryHour    = 9
	feedbackDigestConsistencyLag  = 5 * time.Minute
	feedbackDigestClaimStaleAfter = 2 * time.Hour
	feedbackDigestDateLayout      = "2006-01-02"
	feedbackDigestItemLimit       = 5
	feedbackDigestContextLimit    = 20
)

type feedbackDigestRecipient = feedback.CoreDigestRecipient

type feedbackDigestSubscription = feedback.CoreDigestSubscription

type feedbackDigestBoardWindow struct {
	feedbackDigestSubscription
	WindowStart time.Time
}

type feedbackDigestItem = feedback.CoreDigestItem

// ProcessFeedbackDigestEmail sends due feedback digests at 09:00 in each
// recipient's timezone. A single delivery combines all due boards in a
// workspace, while each board keeps its own delivery cursor.
func ProcessFeedbackDigestEmail(ctx context.Context, store feedback.DigestStore, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessFeedbackDigestEmail")
	defer span.End()

	now := time.Now().UTC()
	afterWorkspaceID := uuid.Nil
	afterUserID := uuid.Nil
	hasCursor := false
	var processingErrors []error

	for {
		if store == nil {
			return errors.New("feedback digest store is unavailable")
		}
		recipients, err := store.ListDigestRecipients(ctx, feedback.CoreDigestRecipientCursor{
			Limit:            feedbackDigestBatchSize,
			HasCursor:        hasCursor,
			AfterWorkspaceID: afterWorkspaceID,
			AfterUserID:      afterUserID,
		})
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get feedback digest recipients: %w", err)
		}
		if len(recipients) == 0 {
			break
		}

		results, batchErr := processGuidanceEmailBatch(ctx, recipients, func(batchCtx context.Context, recipient feedbackDigestRecipient) guidanceEmailBatchResult {
			if processErr := processFeedbackDigestRecipient(batchCtx, store, log, mailerService, copyGenerator, threader, recipient, now); processErr != nil {
				log.Error(batchCtx, "Failed to process feedback digest recipient",
					"recipient_id", recipient.UserID,
					"workspace_id", recipient.WorkspaceID,
					"error", processErr)
				return guidanceEmailBatchResult{Err: processErr}
			}
			return guidanceEmailBatchResult{Processed: true, Sent: true}
		})
		if batchErr != nil {
			span.RecordError(batchErr)
			return fmt.Errorf("feedback digest batch cancelled: %w", batchErr)
		}
		for _, result := range results {
			if result.Err != nil {
				processingErrors = append(processingErrors, result.Err)
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
	store feedback.DigestStore,
	log *logger.Logger,
	mailerService mailer.Service,
	copyGenerator emailcopy.Generator,
	threader emailthread.GuidancePreparer,
	recipient feedbackDigestRecipient,
	now time.Time,
) error {
	subscriptions, err := store.ListDigestSubscriptions(ctx, recipient.UserID, recipient.WorkspaceID)
	if err != nil {
		return fmt.Errorf("get subscriptions for recipient %s: %w", recipient.UserID, err)
	}

	location := feedbackDigestLocation(recipient.Timezone)
	dueSubscriptions := dueFeedbackDigestSubscriptions(now, location, subscriptions)
	if len(dueSubscriptions) == 0 {
		return nil
	}

	localNow := now.In(location)
	localDate := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.UTC)
	windowEnd := now.Add(-feedbackDigestConsistencyLag)
	windowStart := earliestFeedbackDigestWindowStart(dueSubscriptions, windowEnd)
	deliveryID, claimed, err := store.ClaimDigestDelivery(ctx, feedback.CoreDigestDeliveryClaim{
		WorkspaceID: recipient.WorkspaceID,
		RecipientID: recipient.UserID,
		LocalDate:   localDate,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		StaleBefore: now.Add(-feedbackDigestClaimStaleAfter),
	})
	if err != nil {
		return fmt.Errorf("claim delivery for recipient %s: %w", recipient.UserID, err)
	}
	if !claimed {
		return nil
	}

	items, err := getFeedbackDigestItems(ctx, store, recipient, dueSubscriptions, windowEnd)
	if err != nil {
		return failFeedbackDigestDelivery(ctx, store, deliveryID, fmt.Errorf("get digest items: %w", err))
	}

	boardIDs := feedbackDigestBoardIDs(dueSubscriptions)
	if len(items) == 0 {
		if err := completeFeedbackDigestDelivery(ctx, store, deliveryID, recipient, boardIDs, now, windowEnd, feedback.DigestDeliverySkipped, 0); err != nil {
			return failFeedbackDigestDelivery(ctx, store, deliveryID, fmt.Errorf("complete empty digest: %w", err))
		}
		log.Info(ctx, "Skipped empty feedback digest",
			"recipient_id", recipient.UserID,
			"workspace_id", recipient.WorkspaceID,
			"local_date", localNow.Format(feedbackDigestDateLayout))
		return nil
	}

	itemCount := items[0].TotalCount
	if err := sendFeedbackDigestEmail(ctx, log, mailerService, copyGenerator, threader, deliveryID, recipient, items); err != nil {
		return failFeedbackDigestDelivery(ctx, store, deliveryID, err)
	}

	if err := completeFeedbackDigestDelivery(ctx, store, deliveryID, recipient, boardIDs, now, windowEnd, feedback.DigestDeliverySent, itemCount); err != nil {
		return failFeedbackDigestDelivery(ctx, store, deliveryID, fmt.Errorf("complete sent digest: %w", err))
	}

	log.Info(ctx, "Successfully sent feedback digest",
		"recipient_id", recipient.UserID,
		"workspace_id", recipient.WorkspaceID,
		"local_date", localNow.Format(feedbackDigestDateLayout),
		"item_count", itemCount)
	return nil
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

func getFeedbackDigestItems(
	ctx context.Context,
	store feedback.DigestStore,
	recipient feedbackDigestRecipient,
	subscriptions []feedbackDigestBoardWindow,
	windowEnd time.Time,
) ([]feedbackDigestItem, error) {
	boardIDs := make([]uuid.UUID, 0, len(subscriptions))
	windowStarts := make([]time.Time, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		boardIDs = append(boardIDs, subscription.BoardID)
		windowStarts = append(windowStarts, subscription.WindowStart)
	}
	items, err := store.ListDigestItems(ctx, feedback.CoreDigestItemsQuery{
		RecipientID:  recipient.UserID,
		WorkspaceID:  recipient.WorkspaceID,
		BoardIDs:     boardIDs,
		WindowStarts: windowStarts,
		WindowEnd:    windowEnd,
		Limit:        feedbackDigestContextLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("execute feedback digest items query: %w", err)
	}
	return items, nil
}

func feedbackDigestBoardIDs(subscriptions []feedbackDigestBoardWindow) []uuid.UUID {
	boardIDs := make([]uuid.UUID, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		boardIDs = append(boardIDs, subscription.BoardID)
	}
	return boardIDs
}

func completeFeedbackDigestDelivery(
	ctx context.Context,
	store feedback.DigestStore,
	deliveryID uuid.UUID,
	recipient feedbackDigestRecipient,
	boardIDs []uuid.UUID,
	deliveryAt time.Time,
	windowEnd time.Time,
	status feedback.CoreDigestDeliveryStatus,
	itemCount int32,
) error {
	return store.CompleteDigestDelivery(ctx, feedback.CoreDigestDeliveryCompletion{
		DeliveryID:  deliveryID,
		RecipientID: recipient.UserID,
		WorkspaceID: recipient.WorkspaceID,
		BoardIDs:    boardIDs,
		DeliveredAt: deliveryAt,
		WindowEnd:   windowEnd,
		Status:      status,
		ItemCount:   itemCount,
	})
}

func failFeedbackDigestDelivery(ctx context.Context, store feedback.DigestStore, deliveryID uuid.UUID, processingErr error) error {
	updateErr := store.FailDigestDelivery(ctx, deliveryID, processingErr.Error())
	if updateErr != nil {
		return errors.Join(processingErr, fmt.Errorf("mark feedback digest delivery failed: %w", updateErr))
	}
	return processingErr
}

func sendFeedbackDigestEmail(
	ctx context.Context,
	log *logger.Logger,
	mailerService mailer.Service,
	copyGenerator emailcopy.Generator,
	threader emailthread.GuidancePreparer,
	deliveryID uuid.UUID,
	recipient feedbackDigestRecipient,
	items []feedbackDigestItem,
) error {
	if len(items) == 0 {
		return nil
	}

	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", recipient.WorkspaceSlug)
	subject := feedbackDigestSubject(feedbackDigestCount(items[0].TotalCount), recipient.WorkspaceName)
	title := subject
	message := formatFeedbackDigestEmailContent(items, workspaceURL)
	message = appendGuidanceReplyPrompt(message, "I’m Maya, your AI agent. Reply to this email with what these signals change for your product plan or what you want to review next.")
	ctaURL := fmt.Sprintf("%s/teams/%s/feedback/%s", workspaceURL, items[0].TeamID, items[0].ID)
	ctaLabel := "Review latest feedback"

	copyRequest, destinations := feedbackDigestCopyRequest(recipient, items, workspaceURL)
	if copyGenerator != nil {
		generated, err := copyGenerator.Generate(ctx, copyRequest)
		if err != nil {
			log.Info(ctx, "Using deterministic feedback digest copy", "reason", err)
		} else if generatedMessage, renderErr := renderGeneratedEmailContent(generated, destinations); renderErr != nil {
			log.Info(ctx, "Using deterministic feedback digest copy after generated copy could not be rendered", "reason", renderErr)
		} else {
			subject = generated.Subject.Text
			title = generated.H1.Text
			message = generatedMessage
			if generatedLabel, generatedURL, ok := generatedPrimaryCTA(generated, destinations); ok {
				ctaLabel = generatedLabel
				ctaURL = generatedURL
			}
		}
	}

	data := map[string]any{
		"UserName":                 recipient.UserName,
		"UserEmail":                recipient.UserEmail,
		"WorkspaceName":            recipient.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        title,
		"NotificationMessage":      message,
		"NotificationType":         "feedback_digest",
		"NotificationCTAURL":       ctaURL,
		"NotificationCTALabel":     ctaLabel,
		"NotificationsSettingsURL": fmt.Sprintf("%s/settings/workspace/feedback", workspaceURL),
	}
	messageID := feedbackDigestMessageID(deliveryID)
	targets := make([]emailthread.TargetContext, 0, len(items))
	for _, item := range items {
		targets = append(targets, emailthread.TargetContext{
			Kind:        "feedback",
			ID:          item.ID,
			TeamID:      item.TeamID,
			DisplayName: item.Title,
		})
	}
	threadContext, err := emailthread.EncodeThreadContext(emailthread.ThreadContext{
		Source:        "feedback_digest",
		WorkspaceSlug: recipient.WorkspaceSlug,
		Targets:       targets,
	})
	if err != nil {
		return err
	}
	plainText := guidancePlainText(title, message, ctaLabel, ctaURL)
	replyTo, err := prepareGuidanceThread(ctx, threader, emailthread.GuidanceInput{
		WorkspaceID:       recipient.WorkspaceID,
		UserID:            recipient.UserID,
		RecipientEmail:    recipient.UserEmail,
		ExternalThreadID:  messageID,
		InternetMessageID: messageID,
		Subject:           subject,
		Content:           plainText,
		Context:           threadContext,
	})
	if err != nil {
		return fmt.Errorf("prepare feedback digest reply thread: %w", err)
	}

	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:            []string{recipient.UserEmail},
		Template:      "notifications/notification",
		Subject:       subject,
		Data:          data,
		PlainTextBody: plainText,
		Sender:        mailer.SenderProfileMaya,
		ReplyTo:       replyTo,
		MessageID:     messageID,
	}); err != nil {
		return fmt.Errorf("send feedback digest email: %w", err)
	}
	return nil
}

func feedbackDigestCopyRequest(
	recipient feedbackDigestRecipient,
	items []feedbackDigestItem,
	workspaceURL string,
) (emailcopy.Request, map[string]emailCopyDestination) {
	totalCount := feedbackDigestCount(items[0].TotalCount)
	pendingReviewCount := feedbackDigestCount(items[0].PendingReviewCount)
	feedbackItemLabel := pluralize(totalCount, "item", "items")
	pendingVerb := pluralize(pendingReviewCount, "is", "are")
	facts := []emailcopy.Fact{{
		ReferenceID:  "feedback_summary",
		Text:         fmt.Sprintf("%s has %d new feedback %s in this digest; %d %s pending or under review.", recipient.WorkspaceName, totalCount, feedbackItemLabel, pendingReviewCount, pendingVerb),
		EntityTokens: nonEmptyFactTokens(recipient.WorkspaceName),
		ProtectedTokens: nonEmptyFactTokens(
			fmt.Sprintf("%d new feedback %s in this digest", totalCount, feedbackItemLabel),
			fmt.Sprintf("%d %s pending or under review", pendingReviewCount, pendingVerb),
		),
		Required: true,
	}}
	destinations := map[string]emailCopyDestination{
		"feedback_primary": {
			Label: items[0].Title,
			URL:   fmt.Sprintf("%s/teams/%s/feedback/%s", workspaceURL, items[0].TeamID, items[0].ID),
		},
	}

	displayedCount := min(len(items), feedbackDigestItemLimit)
	for index, item := range items {
		referenceID := "feedback_" + item.ID.String()
		factText := fmt.Sprintf(
			"Feedback titled %q was submitted by %s to %s and currently has status %s.",
			item.Title,
			item.AuthorName,
			item.TeamName,
			strings.ReplaceAll(item.Status, "_", " "),
		)
		if description := feedbackDigestPlainText(item.Description, 1200); description != "" {
			factText += " The submitted description is: " + description
		}
		facts = append(facts, emailcopy.Fact{
			ReferenceID:  referenceID,
			Text:         factText,
			EntityTokens: []string{item.Title},
			ProtectedTokens: nonEmptyFactTokens(
				"submitted by "+item.AuthorName,
				"to "+item.TeamName,
				"currently has status "+strings.ReplaceAll(item.Status, "_", " "),
			),
			Required: index < displayedCount,
		})
		if index < displayedCount {
			destinations[referenceID] = emailCopyDestination{
				Label: item.Title,
				URL:   fmt.Sprintf("%s/teams/%s/feedback/%s", workspaceURL, item.TeamID, item.ID),
			}
		}
	}

	return emailcopy.Request{
		SafetyIdentifier:            recipient.UserID.String(),
		Purpose:                     "feedback digest",
		ProductVoice:                "Help the product team understand what customers need, spot grounded patterns, and choose the next feedback to review.",
		Facts:                       facts,
		Actions:                     []emailcopy.Action{{ReferenceID: "feedback_primary", Description: "Open the latest feedback item represented in this digest."}},
		IncludeFeedbackThemeSummary: len(items) >= 3,
		IncludeReplyPrompt:          true,
	}, destinations
}

func feedbackDigestPlainText(value string, maxRunes int) string {
	tokenizer := htmlparser.NewTokenizer(strings.NewReader(value))
	var text strings.Builder
	for {
		switch tokenizer.Next() {
		case htmlparser.ErrorToken:
			plain := strings.Join(strings.Fields(text.String()), " ")
			runes := []rune(plain)
			if maxRunes > 0 && len(runes) > maxRunes {
				return strings.TrimSpace(string(runes[:maxRunes-1])) + "…"
			}
			return plain
		case htmlparser.TextToken:
			if text.Len() > 0 {
				text.WriteByte(' ')
			}
			text.Write(tokenizer.Text())
		}
	}
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

	totalCount := feedbackDigestCount(items[0].TotalCount)
	pendingReviewCount := feedbackDigestCount(items[0].PendingReviewCount)
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

func feedbackDigestCount(value int32) int {
	count, err := safecast.Int64(int64(value))
	if err != nil || count < 0 {
		return 0
	}
	return count
}
