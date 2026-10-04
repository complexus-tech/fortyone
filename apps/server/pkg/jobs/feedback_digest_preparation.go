package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/mailer"
)

// FeedbackDigestBatch is a claimed, immutable section of a workspace summary.
// Its board cursors advance only when the shared summary delivery completes.
type FeedbackDigestBatch struct {
	Section    mailer.Digest
	Targets    []emailthread.TargetContext
	Completion feedback.CoreDigestDeliveryCompletion
}

// PrepareFeedbackDigest retains each board's frequency and current-access
// checks while allowing feedback to share an email with unread workspace work.
func PrepareFeedbackDigest(ctx context.Context, store feedback.DigestStore, recipient feedback.CoreDigestRecipient, now time.Time) (*FeedbackDigestBatch, error) {
	if store == nil {
		return nil, nil
	}
	subscriptions, err := store.ListDigestSubscriptions(ctx, recipient.UserID, recipient.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("load summary feedback subscriptions: %w", err)
	}
	location := feedbackDigestLocation(recipient.Timezone)
	due := dueFeedbackDigestSubscriptions(now, location, subscriptions)
	if len(due) == 0 {
		return nil, nil
	}
	local := now.In(location)
	date := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
	windowEnd := now.Add(-feedbackDigestConsistencyLag)
	id, claimed, err := store.ClaimDigestDelivery(ctx, feedback.CoreDigestDeliveryClaim{
		WorkspaceID: recipient.WorkspaceID, RecipientID: recipient.UserID,
		LocalDate: date, WindowStart: earliestFeedbackDigestWindowStart(due, windowEnd),
		WindowEnd: windowEnd, StaleBefore: now.Add(-feedbackDigestClaimStaleAfter),
	})
	if err != nil || !claimed {
		return nil, err
	}
	items, err := getFeedbackDigestItems(ctx, store, recipient, due, windowEnd)
	if err != nil {
		return nil, failFeedbackDigestDelivery(ctx, store, id, err)
	}
	completion := feedback.CoreDigestDeliveryCompletion{
		DeliveryID: id, RecipientID: recipient.UserID, WorkspaceID: recipient.WorkspaceID,
		BoardIDs: feedbackDigestBoardIDs(due), DeliveredAt: now, WindowEnd: windowEnd,
		Status: feedback.DigestDeliverySent,
	}
	if len(items) == 0 {
		completion.Status = feedback.DigestDeliverySkipped
		if err := store.CompleteDigestDelivery(ctx, completion); err != nil {
			return nil, failFeedbackDigestDelivery(ctx, store, id, err)
		}
		return nil, nil
	}
	completion.ItemCount = items[0].TotalCount
	base := "https://" + recipient.WorkspaceSlug + ".fortyone.app"
	batch := &FeedbackDigestBatch{
		Completion: completion,
		Section:    mailer.Digest{Intro: fmt.Sprintf("%d new feedback %s to review.", completion.ItemCount, pluralize(int(completion.ItemCount), "item", "items"))},
	}
	for _, item := range items[:min(len(items), mailer.DigestDetailLimit)] {
		url := fmt.Sprintf("%s/teams/%s/feedback/%s", base, item.TeamID, item.ID)
		batch.Section.Rows = append(batch.Section.Rows, mailer.DigestRow{
			Label: item.Title, Text: fmt.Sprintf("%s submitted this to %s · %s", item.AuthorName, item.TeamName, strings.ReplaceAll(item.Status, "_", " ")),
			Detail: feedbackDigestPlainText(item.Description, 180), URL: url, Icon: "comment",
		})
		batch.Targets = append(batch.Targets, emailthread.TargetContext{Kind: "feedback", ID: item.ID, TeamID: item.TeamID, DisplayName: item.Title})
	}
	batch.Section.Rows = appendBriefingMore(batch.Section.Rows, int(completion.ItemCount), base, "feedback items")
	return batch, nil
}
