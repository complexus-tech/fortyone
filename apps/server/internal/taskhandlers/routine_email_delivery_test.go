package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type summaryFeedbackStore struct {
	feedback.DigestStore
	subscriptions []feedback.CoreDigestSubscription
	items         []feedback.CoreDigestItem
	claim         feedback.CoreDigestDeliveryClaim
	claimID       uuid.UUID
	completions   []feedback.CoreDigestDeliveryCompletion
	failures      int
}

func (s *summaryFeedbackStore) ListDigestSubscriptions(context.Context, uuid.UUID, uuid.UUID) ([]feedback.CoreDigestSubscription, error) {
	return s.subscriptions, nil
}
func (s *summaryFeedbackStore) ClaimDigestDelivery(_ context.Context, claim feedback.CoreDigestDeliveryClaim) (uuid.UUID, bool, error) {
	s.claim = claim
	s.claimID = uuid.New()
	return s.claimID, true, nil
}
func (s *summaryFeedbackStore) ListDigestItems(context.Context, feedback.CoreDigestItemsQuery) ([]feedback.CoreDigestItem, error) {
	return s.items, nil
}
func (s *summaryFeedbackStore) CompleteDigestDelivery(_ context.Context, completion feedback.CoreDigestDeliveryCompletion) error {
	s.completions = append(s.completions, completion)
	return nil
}
func (s *summaryFeedbackStore) FailDigestDelivery(context.Context, uuid.UUID, string) error {
	s.failures++
	return nil
}

func summaryFixture(t *testing.T, now time.Time) (*handlers, *routineStoreStub, *notificationDeliveryStoreStub, *summaryFeedbackStore, *briefingMailerStub, *asynq.Task) {
	t.Helper()
	recipient := notifications.RoutineRecipient{UserID: uuid.New(), WorkspaceID: uuid.New(), Email: "reviewer@example.com", Name: "Reviewer", WorkspaceSlug: "product", WorkspaceName: "Product", Timezone: "Africa/Harare"}
	routine := &routineStoreStub{recipient: &recipient}
	delivery := &notificationDeliveryStoreStub{digest: &notifications.EmailDigest{
		RecipientID: recipient.UserID, WorkspaceID: recipient.WorkspaceID, UserEmail: recipient.Email,
		WorkspaceSlug: recipient.WorkspaceSlug, WorkspaceName: recipient.WorkspaceName, WorkspaceRole: "member",
		Items: []notifications.EmailDigestItem{{NotificationID: uuid.New(), EntityID: uuid.New(), EntityType: notifications.EntityTypeStory,
			NotificationType: notifications.NotificationTypeStoryUpdate, ContentHash: []byte("observed-version"), Title: "Ship the update",
			Message: json.RawMessage(`{"template":"Sam changed priority to High."}`), CreatedAt: now.Add(-time.Hour)}},
	}}
	feedbackStore := &summaryFeedbackStore{
		subscriptions: []feedback.CoreDigestSubscription{{BoardID: uuid.New(), TeamID: uuid.New(), EmailFrequency: "daily", CreatedAt: now.Add(-24 * time.Hour)}},
		items:         []feedback.CoreDigestItem{{ID: uuid.New(), TeamID: uuid.New(), Title: "Improve search", AuthorName: "Alex", TeamName: "Product", Status: "new", TotalCount: 1}},
	}
	sender := &briefingMailerStub{}
	sources := &briefingStoreStub{}
	h := &handlers{log: logger.NewWithText(io.Discard, slog.LevelError, "summary-test"), routineDeliveries: routine,
		notificationDeliveries: delivery, feedbackDigest: feedbackStore, mailerService: sender,
		briefingSources: jobs.BriefingSources{Stories: sources, Objectives: sources, Weekly: sources}}
	payload, err := json.Marshal(tasks.NotificationEmailDigestPayload{RecipientID: recipient.UserID, WorkspaceID: recipient.WorkspaceID})
	require.NoError(t, err)
	return h, routine, delivery, feedbackStore, sender, asynq.NewTask(tasks.TypeNotificationEmailDigest, payload)
}

func TestRoutineSummaryCombinesActivityAndFeedbackWithAtomicCompletion(t *testing.T) {
	now := time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC)
	h, routine, delivery, feedbackStore, sender, task := summaryFixture(t, now)
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now))
	require.Len(t, sender.emails, 1)
	sections := sender.emails[0].Data.(map[string]any)["NotificationSections"].([]mailer.Digest)
	require.Len(t, sections, 2)
	require.Contains(t, sender.emails[0].PlainTextBody, "High")
	require.Contains(t, sender.emails[0].PlainTextBody, "Improve search")
	require.Len(t, routine.completions, 1)
	completion := routine.completions[0]
	require.True(t, completion.Sent)
	require.Equal(t, delivery.digest.Items[0].ContentHash, completion.NotificationSnapshots[0].ContentHash)
	require.Equal(t, feedbackStore.claimID, completion.FeedbackDigest.DeliveryID)
	require.Empty(t, feedbackStore.completions, "feedback cursors must join the routine adapter transaction")
	// Changed activity remains pending after the weekly send.
	delivery.digest.Items[0].ContentHash = []byte("newer-version")
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now.Add(time.Hour)))
	require.Len(t, sender.emails, 1)
	for day := 1; day < 7; day++ {
		require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now.AddDate(0, 0, day)))
	}
	require.Len(t, sender.emails, 1)
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now.AddDate(0, 0, 7)))
	require.Len(t, sender.emails, 2)
	require.NotEqual(t, sender.emails[0].MessageID, sender.emails[1].MessageID)
}

func TestRoutineSummaryFeedbackWithoutActivity(t *testing.T) {
	now := time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC)
	h, routine, delivery, _, sender, task := summaryFixture(t, now)
	delivery.digest = nil
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now))
	require.Len(t, sender.emails, 1)
	require.Equal(t, []string{"reviewer@example.com"}, sender.emails[0].To)
	require.Contains(t, sender.emails[0].PlainTextBody, "Improve search")
	require.NotContains(t, sender.emails[0].PlainTextBody, "0 updates")
	require.Empty(t, routine.completions[0].NotificationSnapshots)
	require.NotNil(t, routine.completions[0].FeedbackDigest)
}

func TestRoutineSummaryFailureReleasesBothClaimsWithoutAdvancing(t *testing.T) {
	now := time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC)
	h, routine, _, feedbackStore, sender, task := summaryFixture(t, now)
	sender.err = errors.New("SMTP unavailable")
	require.Error(t, h.handleNotificationEmailDigestAt(t.Context(), task, now))
	require.Empty(t, routine.completions)
	require.Empty(t, feedbackStore.completions)
	require.Equal(t, 1, routine.failures)
	require.Equal(t, 1, feedbackStore.failures)
}

func TestRoutineSummaryCompletionFailureDoesNotReleaseAcceptedEmail(t *testing.T) {
	now := time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC)
	h, routine, _, feedbackStore, sender, task := summaryFixture(t, now)
	routine.completeErr = errors.New("completion unavailable")
	require.Error(t, h.handleNotificationEmailDigestAt(t.Context(), task, now))
	require.Len(t, sender.emails, 1)
	require.Zero(t, routine.failures)
	require.Zero(t, feedbackStore.failures)
	// A stale retry after SMTP acceptance cannot reopen this week's allowance.
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now.Add(11*time.Minute)))
	require.Len(t, sender.emails, 1)
}

func TestRoutineSummaryBeforeNineAndEmptyCheckDoNotConsumeWeeklySend(t *testing.T) {
	now := time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC)
	h, routine, delivery, feedbackStore, sender, task := summaryFixture(t, now)
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now.Add(-2*time.Hour)))
	require.Empty(t, routine.claims)
	require.Empty(t, sender.emails)
	items := delivery.digest.Items
	delivery.digest.Items, feedbackStore.items = nil, nil
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now))
	require.False(t, routine.completions[0].Sent)
	require.Empty(t, routine.sentWeeks)
	delivery.digest.Items = items
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now.Add(time.Hour)))
	require.Len(t, sender.emails, 1)
}

func TestRoutineSummaryFencesExpiredOwnerBeforeSMTP(t *testing.T) {
	now := time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC)
	h, routine, _, feedbackStore, sender, task := summaryFixture(t, now)
	routine.beginErr = errors.New("claim expired")
	require.Error(t, h.handleNotificationEmailDigestAt(t.Context(), task, now))
	require.Empty(t, sender.emails)
	require.Empty(t, routine.completions)
	require.Equal(t, 1, feedbackStore.failures)
}

func TestWeeklySummaryCatchUpIncludesPlanningWithCurrentFacts(t *testing.T) {
	now := time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC) // Tuesday, 10 a.m. in Harare.
	h, routine, _, _, sender, task := summaryFixture(t, now)
	routine.recipient.WeeklyEnabled = true
	sources := &briefingStoreStub{stats: notifications.WeeklyDigestStats{OverdueStories: 2}}
	h.briefingSources = jobs.BriefingSources{Stories: sources, Objectives: sources, Weekly: sources}
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now))
	require.Len(t, sender.emails, 1)
	require.Contains(t, sender.emails[0].PlainTextBody, "Your week ahead")
	require.Equal(t, time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC), sources.weeklyQuery.AsOf)
	require.Equal(t, time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC), *routine.completions[0].GuidanceDate)
}
