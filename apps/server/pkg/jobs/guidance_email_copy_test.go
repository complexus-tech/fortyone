package jobs

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type recordingEmailCopyGenerator struct {
	output   emailcopy.Output
	err      error
	requests []emailcopy.Request
}

func (g *recordingEmailCopyGenerator) Generate(_ context.Context, request emailcopy.Request) (emailcopy.Output, error) {
	g.requests = append(g.requests, request)
	return g.output, g.err
}

type recordingMailer struct {
	templatedEmails []mailer.TemplatedEmail
}

type recordingGuidancePreparer struct {
	inputs []emailthread.GuidanceInput
	err    error
}

func (preparer *recordingGuidancePreparer) PrepareGuidance(_ context.Context, input emailthread.GuidanceInput) (emailthread.PreparedGuidance, error) {
	preparer.inputs = append(preparer.inputs, input)
	return emailthread.PreparedGuidance{
		Thread:  messaging.EmailThreadRecord{ID: uuid.New()},
		ReplyTo: "maya+opaque-token@reply.fortyone.app",
	}, preparer.err
}

func (m *recordingMailer) Send(_ context.Context, _ mailer.Email) error {
	return nil
}

func (m *recordingMailer) SendTemplated(_ context.Context, email mailer.TemplatedEmail) error {
	m.templatedEmails = append(m.templatedEmails, email)
	return nil
}

func TestSendWeeklyDigestEmailUsesCompleteGeneratedCopyAndMayaSender(t *testing.T) {
	recipient := WeeklyDigestRecipient{
		UserID:        uuid.New(),
		UserEmail:     "joseph@example.com",
		UserName:      "Joseph",
		WorkspaceID:   uuid.New(),
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
	}
	stats := WeeklyDigestStats{UnreadNotifications: 2, UnreadPriorityNotifications: 1, OverdueStories: 3}
	generator := &recordingEmailCopyGenerator{output: emailcopy.Output{
		Subject: emailcopy.GroundedText{Text: "Your week has two clear priorities", ReferenceIDs: []string{"unread_updates", "overdue_tasks"}},
		H1:      emailcopy.GroundedText{Text: "Move the most important work forward", ReferenceIDs: []string{"unread_updates", "overdue_tasks"}},
		Intro:   emailcopy.GroundedText{Text: "Start with the conversations and commitments most likely to unblock progress.", ReferenceIDs: []string{"unread_updates", "overdue_tasks"}},
		Rows: []emailcopy.Row{
			{ReferenceID: "unread_updates", Text: "Review 2 unread updates, including 1 mention or reply."},
			{ReferenceID: "overdue_tasks", Text: "Decide how to move 3 overdue assigned tasks forward."},
		},
		CTAs: []emailcopy.CTA{{ReferenceID: "review_week", Label: "Shape the week"}},
	}}
	mailerService := &recordingMailer{}

	err := sendWeeklyDigestEmail(context.Background(), newTestJobLogger(), mailerService, generator, nil, recipient, stats)
	require.NoError(t, err)
	require.Len(t, generator.requests, 1)
	require.Len(t, mailerService.templatedEmails, 1)

	request := generator.requests[0]
	require.Equal(t, mayaGuidanceProductVoice, request.ProductVoice)
	require.Contains(t, factTextByReference(request.Facts, "unread_updates"), "2 unread updates, including 1 mention or reply")
	require.NotContains(t, request.Actions[0].Description, "https://")

	email := mailerService.templatedEmails[0]
	require.Equal(t, mailer.SenderProfileMaya, email.Sender)
	require.Equal(t, "Your week has two clear priorities", email.Subject)
	data := requireEmailData(t, email)
	require.Equal(t, "Move the most important work forward", data["NotificationTitle"])
	require.Equal(t, "Shape the week", data["NotificationCTALabel"])
	require.Equal(t, "https://product.fortyone.app/my-work?tab=assigned", data["NotificationCTAURL"])
	require.Contains(t, data["NotificationMessage"], "Start with the conversations")
	require.Contains(t, data["NotificationMessage"], "3 overdue assigned tasks")
}

func TestSendWeeklyDigestEmailFallsBackWhenGenerationFails(t *testing.T) {
	recipient := WeeklyDigestRecipient{
		UserID:        uuid.New(),
		UserEmail:     "joseph@example.com",
		UserName:      "Joseph",
		WorkspaceID:   uuid.New(),
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
	}
	generator := &recordingEmailCopyGenerator{err: errors.New("generation timed out")}
	mailerService := &recordingMailer{}

	err := sendWeeklyDigestEmail(context.Background(), newTestJobLogger(), mailerService, generator, nil, recipient, WeeklyDigestStats{OverdueStories: 2})
	require.NoError(t, err)
	require.Len(t, mailerService.templatedEmails, 1)

	email := mailerService.templatedEmails[0]
	require.Equal(t, mailer.SenderProfileMaya, email.Sender)
	require.Equal(t, "Weekly digest: Product", email.Subject)
	data := requireEmailData(t, email)
	require.Equal(t, "Plan my week", data["NotificationCTALabel"])
	require.Contains(t, data["NotificationMessage"], "2 overdue assigned tasks")
}

func TestSendWeeklyDigestEmailCreatesReplyThreadAndMultipartMessage(t *testing.T) {
	recipient := WeeklyDigestRecipient{
		UserID:        uuid.New(),
		UserEmail:     "joseph@example.com",
		UserName:      "Joseph",
		WorkspaceID:   uuid.New(),
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
	}
	mailerService := &recordingMailer{}
	threader := &recordingGuidancePreparer{}

	err := sendWeeklyDigestEmail(context.Background(), newTestJobLogger(), mailerService, nil, threader, recipient, WeeklyDigestStats{OverdueStories: 1})
	require.NoError(t, err)
	require.Len(t, threader.inputs, 1)
	require.Len(t, mailerService.templatedEmails, 1)

	input := threader.inputs[0]
	require.Equal(t, recipient.WorkspaceID, input.WorkspaceID)
	require.Equal(t, recipient.UserID, input.UserID)
	require.Contains(t, string(input.Context), `"source":"weekly_digest"`)
	require.NotEmpty(t, input.Content)

	email := mailerService.templatedEmails[0]
	require.Equal(t, "maya+opaque-token@reply.fortyone.app", email.ReplyTo)
	require.NotEmpty(t, email.PlainTextBody)
	require.Contains(t, email.PlainTextBody, "I’m Maya, your AI agent")
	require.Contains(t, email.PlainTextBody, "Reply to this email")
	require.Equal(t, input.InternetMessageID, email.MessageID)
}

func TestSendOverdueStoriesEmailLinksGeneratedRowsToCanonicalTasks(t *testing.T) {
	story := OverdueStory{
		ID:             uuid.New(),
		Title:          "Confirm launch scope",
		EndDate:        time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC),
		AssigneeID:     uuid.New(),
		AssigneeEmail:  "joseph@example.com",
		AssigneeName:   "Joseph",
		WorkspaceID:    uuid.New(),
		WorkspaceName:  "Product",
		WorkspaceSlug:  "product",
		TeamName:       "Launch",
		TeamCode:       "PRD",
		SequenceID:     571,
		StatusName:     "In progress",
		DeadlineStatus: "overdue",
		DaysDifference: 3,
	}
	storyReference := "task:" + story.ID.String()
	generator := &recordingEmailCopyGenerator{output: emailcopy.Output{
		Subject: emailcopy.GroundedText{Text: "One commitment needs a decision", ReferenceIDs: []string{"task_summary"}},
		H1:      emailcopy.GroundedText{Text: "Protect the launch commitment", ReferenceIDs: []string{storyReference}},
		Intro:   emailcopy.GroundedText{Text: "A quick update will give the team a reliable next step.", ReferenceIDs: []string{storyReference}},
		Rows: []emailcopy.Row{
			{ReferenceID: "task_summary", Text: "There is 1 assigned task that needs attention."},
			{ReferenceID: storyReference, Text: "Confirm launch scope is 3 days overdue; confirm the owner or reset the date."},
		},
		CTAs: []emailcopy.CTA{{ReferenceID: "review_assigned_work", Label: "Resolve the commitment"}},
	}}
	mailerService := &recordingMailer{}

	err := sendOverdueStoriesEmailForAssignee(context.Background(), newTestJobLogger(), mailerService, generator, nil, []OverdueStory{story})
	require.NoError(t, err)
	require.Len(t, generator.requests, 1)

	request := generator.requests[0]
	require.Contains(t, factTextByReference(request.Facts, storyReference), "3 days overdue")
	require.Contains(t, factTextByReference(request.Facts, storyReference), "August 9, 2026")
	require.NotContains(t, factTextByReference(request.Facts, storyReference), "https://")

	email := mailerService.templatedEmails[0]
	require.Equal(t, mailer.SenderProfileMaya, email.Sender)
	require.Equal(t, "One commitment needs a decision", email.Subject)
	data := requireEmailData(t, email)
	require.Equal(t, "Protect the launch commitment", data["NotificationTitle"])
	require.Contains(t, data["NotificationMessage"], `href="https://product.fortyone.app/work/PRD-571"`)
	require.Contains(t, data["NotificationMessage"], ">Confirm launch scope</a>")
	require.Equal(t, "Resolve the commitment", data["NotificationCTALabel"])
}

func TestOverdueStoriesEmailCopyRequestLimitsRowsAndAccountsForRemainder(t *testing.T) {
	stories := make([]OverdueStory, 15)
	for index := range stories {
		stories[index] = OverdueStory{
			ID:             uuid.New(),
			Title:          "Task " + strings.Repeat("x", index+1),
			EndDate:        time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC),
			WorkspaceName:  "Product",
			DeadlineStatus: "overdue",
			DaysDifference: 3,
		}
	}

	request, destinations := overdueStoriesEmailCopyRequest(stories, "https://product.fortyone.app", "https://product.fortyone.app/my-work?tab=assigned")

	require.Len(t, request.Facts, maxGuidanceEmailRows+1) // Optional workspace context plus twelve required rows.
	require.Contains(t, factTextByReference(request.Facts, "task_summary"), "15 assigned tasks")
	require.Contains(t, factTextByReference(request.Facts, "task_summary"), "11 of them")
	require.Contains(t, factTextByReference(request.Facts, "task_summary"), "4 more")
	require.Len(t, destinations, maxGuidanceEmailRows) // Eleven task links plus the primary CTA.
}

func TestSendObjectiveOverdueEmailLinksObjectivesAndKeyResults(t *testing.T) {
	teamID := uuid.New()
	objectiveID := uuid.New()
	keyResultID := uuid.New()
	keyResults := `[{"id":"` + keyResultID.String() + `","name":"Lift activation","end_date":"2026-08-10","measurement_type":"percentage","current_value":15,"target_value":40,"is_completed":false,"deadline_status":"overdue","days_difference":2}]`
	objective := OverdueObjective{
		ID:             objectiveID,
		Name:           "Improve onboarding",
		EndDate:        time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC),
		LeadUserID:     uuid.New(),
		LeadEmail:      "joseph@example.com",
		LeadName:       "Joseph",
		WorkspaceID:    uuid.New(),
		WorkspaceName:  "Product",
		WorkspaceSlug:  "product",
		TeamID:         teamID,
		DeadlineStatus: "future",
		KeyResults:     keyResults,
	}
	objectiveReference := "objective:" + objectiveID.String()
	keyResultReference := "key_result:" + keyResultID.String()
	generator := &recordingEmailCopyGenerator{output: emailcopy.Output{
		Subject: emailcopy.GroundedText{Text: "Onboarding progress needs one adjustment", ReferenceIDs: []string{objectiveReference, keyResultReference}},
		H1:      emailcopy.GroundedText{Text: "Keep onboarding on course", ReferenceIDs: []string{objectiveReference, keyResultReference}},
		Intro:   emailcopy.GroundedText{Text: "The objective is on schedule, but one result needs a credible recovery plan.", ReferenceIDs: []string{objectiveReference, keyResultReference}},
		Rows: []emailcopy.Row{
			{ReferenceID: "objective_summary", Text: "1 objective has 2 signals that need attention."},
			{ReferenceID: objectiveReference, Text: "Improve onboarding is on schedule; keep its outcome visible while the result recovers."},
			{ReferenceID: keyResultReference, Text: "Lift activation is at 15% against a 40% target and is 2 days overdue."},
		},
		CTAs: []emailcopy.CTA{{ReferenceID: "review_objectives", Label: "Build the recovery plan"}},
	}}
	mailerService := &recordingMailer{}

	err := sendObjectiveOverdueEmailForLead(context.Background(), newTestJobLogger(), mailerService, generator, nil, []OverdueObjective{objective})
	require.NoError(t, err)
	require.Len(t, generator.requests, 1)

	request := generator.requests[0]
	require.Contains(t, factTextByReference(request.Facts, keyResultReference), "15% against a 40% target")
	require.Contains(t, factTextByReference(request.Facts, keyResultReference), "2 days overdue")
	require.NotContains(t, factTextByReference(request.Facts, keyResultReference), "https://")

	email := mailerService.templatedEmails[0]
	require.Equal(t, mailer.SenderProfileMaya, email.Sender)
	data := requireEmailData(t, email)
	objectiveURL := "https://product.fortyone.app/teams/" + teamID.String() + "/objectives/" + objectiveID.String()
	require.Contains(t, data["NotificationMessage"], `href="`+objectiveURL+`"`)
	require.Contains(t, data["NotificationMessage"], `href="`+objectiveURL+`?tab=overview&amp;keyResultId=`+keyResultID.String()+`"`)
	require.Equal(t, "Build the recovery plan", data["NotificationCTALabel"])
	require.Equal(t, "https://product.fortyone.app/roadmap", data["NotificationCTAURL"])
}

func newTestJobLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelDebug, "jobs-test")
}

func requireEmailData(t *testing.T, email mailer.TemplatedEmail) map[string]any {
	t.Helper()
	data, ok := email.Data.(map[string]any)
	require.True(t, ok)
	return data
}

func factTextByReference(facts []emailcopy.Fact, referenceID string) string {
	for _, fact := range facts {
		if fact.ReferenceID == referenceID {
			return fact.Text
		}
	}
	return ""
}
