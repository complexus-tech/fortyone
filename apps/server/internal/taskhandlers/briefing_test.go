package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type briefingStoreStub struct {
	guidanceStoreStub
	stories    []stories.OverdueGuidanceStory
	objectives []objectives.OverdueGuidanceObjective
	stats      notifications.WeeklyDigestStats
}

func (s *briefingStoreStub) ListOverdueStoryGuidanceItems(context.Context, time.Time, uuid.UUID, uuid.UUID) ([]stories.OverdueGuidanceStory, error) {
	return s.stories, nil
}
func (s *briefingStoreStub) ListOverdueObjectiveGuidanceItems(context.Context, time.Time, uuid.UUID, uuid.UUID) ([]objectives.OverdueGuidanceObjective, error) {
	return s.objectives, nil
}
func (s *briefingStoreStub) GetWeeklyDigestStats(context.Context, notifications.WeeklyDigestStatsQuery) (notifications.WeeklyDigestStats, error) {
	return s.stats, nil
}

type routineStoreStub struct {
	RoutineDeliveryStore
	done        bool
	completions []notifications.RoutineCompletion
	failures    int
	completeErr error
	recipient   *notifications.RoutineRecipient
}

func (s *routineStoreStub) ClaimRoutine(context.Context, notifications.RoutineClaim) (uuid.UUID, error) {
	if s.done {
		return uuid.Nil, nil
	}
	return uuid.New(), nil
}
func (s *routineStoreStub) CompleteRoutine(_ context.Context, c notifications.RoutineCompletion) error {
	s.done = true
	s.completions = append(s.completions, c)
	return s.completeErr
}
func (s *routineStoreStub) FailRoutine(context.Context, uuid.UUID) error { s.failures++; return nil }

func (s *routineStoreStub) GetRoutineRecipient(context.Context, notifications.DeliveryScope) (*notifications.RoutineRecipient, error) {
	return s.recipient, nil
}
func (s *routineStoreStub) HasRoutineGuidance(_ context.Context, _ notifications.DeliveryScope, date time.Time) (bool, error) {
	for _, c := range s.completions {
		if c.Sent && c.GuidanceDate != nil && c.GuidanceDate.Equal(date) {
			return true, nil
		}
	}
	return false, nil
}

type briefingMailerStub struct {
	guidanceMailerStub
	emails []mailer.TemplatedEmail
	err    error
}

func (s *briefingMailerStub) SendTemplated(_ context.Context, email mailer.TemplatedEmail) error {
	s.emails = append(s.emails, email)
	return s.err
}

func TestLegacyNotificationTaskSendsOneBatchAndCoversAllTenItems(t *testing.T) {
	recipient, workspace := uuid.New(), uuid.New()
	store := &notificationDeliveryStoreStub{digest: &notifications.EmailDigest{RecipientID: recipient, WorkspaceID: workspace, UserEmail: "person@example.com", WorkspaceSlug: "product", WorkspaceName: "Product", WorkspaceRole: "admin"}}
	for i := range 10 {
		store.digest.Items = append(store.digest.Items, notifications.EmailDigestItem{NotificationID: uuid.New(), EntityID: uuid.New(), EntityType: notifications.EntityTypeStory, NotificationType: notifications.NotificationTypeStoryUpdate, Title: fmt.Sprintf("Story %d", i+1), Message: json.RawMessage(`{"template":"Sam Taylor updated this story."}`), ActorName: "Sam Taylor", CreatedAt: time.Now().Add(-time.Hour)})
	}
	routine, sender := &routineStoreStub{}, &briefingMailerStub{}
	h := &handlers{log: logger.NewWithText(io.Discard, slog.LevelError, "test"), notificationDeliveries: store, routineDeliveries: routine, mailerService: sender}
	payload, err := json.Marshal(tasks.NotificationEmailPayload{RecipientID: recipient, WorkspaceID: workspace, NotificationID: store.digest.Items[0].NotificationID})
	require.NoError(t, err)
	require.NoError(t, h.HandleNotificationEmail(t.Context(), asynq.NewTask(tasks.TypeNotificationEmail, payload)))
	require.Len(t, sender.emails, 1)
	digest := sender.emails[0].Data.(map[string]any)["NotificationDigest"].(mailer.Digest)
	require.Len(t, digest.Rows, 6)
	require.True(t, digest.Rows[5].More)
	require.Len(t, routine.completions[0].NotificationIDs, 10)
	require.NoError(t, h.HandleNotificationEmail(t.Context(), asynq.NewTask(tasks.TypeNotificationEmail, payload)))
	require.Len(t, sender.emails, 1)
}

func TestActivityCombinesGuidanceOncePerLocalDay(t *testing.T) {
	now := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)
	recipient := notifications.RoutineRecipient{UserID: uuid.New(), WorkspaceID: uuid.New(), WorkspaceSlug: "product", WorkspaceName: "Product", Email: "person@example.com", Timezone: "Africa/Harare"}
	sources := &briefingStoreStub{}
	for i := range 8 {
		sources.stories = append(sources.stories, stories.OverdueGuidanceStory{ID: uuid.New(), Title: fmt.Sprintf("Priority %d", i), EndDate: now, DeadlineStatus: "due_today"})
	}
	delivery := &notificationDeliveryStoreStub{digest: &notifications.EmailDigest{RecipientID: recipient.UserID, WorkspaceID: recipient.WorkspaceID, UserEmail: recipient.Email, WorkspaceSlug: recipient.WorkspaceSlug, WorkspaceName: recipient.WorkspaceName, WorkspaceRole: "admin"}}
	for i := range 10 {
		delivery.digest.Items = append(delivery.digest.Items, notifications.EmailDigestItem{NotificationID: uuid.New(), EntityID: uuid.New(), EntityType: notifications.EntityTypeStory, NotificationType: notifications.NotificationTypeStoryUpdate, Title: fmt.Sprintf("Update %d", i), Message: json.RawMessage(`{"template":"Sam Taylor updated this story."}`), ActorName: "Sam Taylor", CreatedAt: now.Add(-time.Hour)})
	}
	routine, sender := &routineStoreStub{recipient: &recipient}, &briefingMailerStub{}
	h := &handlers{log: logger.NewWithText(io.Discard, slog.LevelError, "test"), routineDeliveries: routine, notificationDeliveries: delivery, mailerService: sender, briefingSources: jobs.BriefingSources{Stories: sources, Objectives: sources, Weekly: sources}}
	payload, err := json.Marshal(tasks.NotificationEmailDigestPayload{RecipientID: recipient.UserID, WorkspaceID: recipient.WorkspaceID})
	require.NoError(t, err)
	task := asynq.NewTask(tasks.TypeNotificationEmailDigest, payload)
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now))
	require.Len(t, sender.emails, 1)
	require.Equal(t, mailer.SenderProfileMaya, sender.emails[0].Sender)
	sections := sender.emails[0].Data.(map[string]any)["NotificationSections"].([]mailer.Digest)
	require.Len(t, sections, 2)
	for _, section := range sections {
		require.Len(t, section.Rows, 6)
		require.True(t, section.Rows[5].More)
	}
	require.Contains(t, sections[0].Rows[5].Text, "3 more stories")
	require.Contains(t, sender.emails[0].PlainTextBody, "Priority 0")
	require.Len(t, routine.completions[0].NotificationIDs, 10)
	require.NotNil(t, routine.completions[0].GuidanceDate)
	// A later batch still sends new activity, without repeating today's priorities.
	routine.done = false
	delivery.digest.Items = delivery.digest.Items[:1]
	delivery.digest.Items[0].NotificationID = uuid.New()
	require.NoError(t, h.handleNotificationEmailDigestAt(t.Context(), task, now.Add(2*time.Hour)))
	require.Len(t, sender.emails, 2)
	require.NotContains(t, sender.emails[1].Data.(map[string]any), "NotificationSections")
	require.Nil(t, routine.completions[1].GuidanceDate)
	// Retired scheduled jobs cannot produce an additional briefing or weekly note.
	require.NoError(t, h.HandleMorningBriefing(t.Context(), asynq.NewTask(tasks.TypeMorningBriefing, nil)))
	require.Len(t, sender.emails, 2)
}

func TestActivitySendFailureDoesNotCoverNotificationsOrGuidance(t *testing.T) {
	recipient, workspace := uuid.New(), uuid.New()
	store := &notificationDeliveryStoreStub{digest: &notifications.EmailDigest{RecipientID: recipient, WorkspaceID: workspace, UserEmail: "person@example.com", WorkspaceSlug: "product", Items: []notifications.EmailDigestItem{{NotificationID: uuid.New(), EntityID: uuid.New(), EntityType: notifications.EntityTypeStory, NotificationType: notifications.NotificationTypeStoryUpdate, Title: "Launch", Message: json.RawMessage(`{"template":"Story updated."}`)}}}}
	routine, sender := &routineStoreStub{}, &briefingMailerStub{err: errors.New("SMTP unavailable")}
	h := &handlers{log: logger.NewWithText(io.Discard, slog.LevelError, "test"), routineDeliveries: routine, notificationDeliveries: store, mailerService: sender}
	payload, err := json.Marshal(tasks.NotificationEmailDigestPayload{RecipientID: recipient, WorkspaceID: workspace})
	require.NoError(t, err)
	require.Error(t, h.HandleNotificationEmailDigest(t.Context(), asynq.NewTask(tasks.TypeNotificationEmailDigest, payload)))
	require.Empty(t, routine.completions)
	require.Equal(t, 1, routine.failures)
}

func TestGuidanceUsesRecipientDate(t *testing.T) {
	for _, tc := range []struct {
		now, zone, date string
		eligible        bool
	}{
		{"2026-09-06T23:00:00Z", "Africa/Harare", "2026-09-07", true},
		{"2026-09-06T23:00:00Z", "America/New_York", "2026-09-06", false},
		{"2026-03-09T13:00:00Z", "America/New_York", "2026-03-09", true},
		{"2026-09-07T03:30:00Z", "Asia/Kolkata", "2026-09-07", true},
		{"2026-09-07T09:00:00Z", "invalid", "2026-09-07", true},
	} {
		now, err := time.Parse(time.RFC3339, tc.now)
		require.NoError(t, err)
		date, eligible := guidanceDate(now, tc.zone)
		require.Equal(t, tc.date, date.Format("2006-01-02"))
		require.Equal(t, tc.eligible, eligible)
	}
}
