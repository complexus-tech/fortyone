package jobs

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFeedbackDigestLocationFallsBackToUTC(t *testing.T) {
	require.Equal(t, time.UTC, feedbackDigestLocation("not/a-timezone"))
	require.Equal(t, time.UTC, feedbackDigestLocation(""))
	require.Equal(t, "Africa/Harare", feedbackDigestLocation("Africa/Harare").String())
}

func TestDueFeedbackDigestSubscriptionsWaitsUntilLocalNineAndCatchesUp(t *testing.T) {
	location := feedbackDigestLocation("Africa/Harare")
	createdAt := time.Date(2026, time.July, 19, 7, 30, 0, 0, time.UTC)
	subscription := feedbackDigestSubscription{
		BoardID:        uuid.New(),
		EmailFrequency: "daily",
		CreatedAt:      createdAt,
	}

	beforeNine := time.Date(2026, time.July, 20, 6, 59, 0, 0, time.UTC)
	require.Empty(t, dueFeedbackDigestSubscriptions(beforeNine, location, []feedbackDigestSubscription{subscription}))

	afterNine := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	due := dueFeedbackDigestSubscriptions(afterNine, location, []feedbackDigestSubscription{subscription})
	require.Len(t, due, 1)
	require.Equal(t, createdAt, due[0].WindowStart)
}

func TestDueFeedbackDigestSubscriptionsOnlySendsOncePerLocalDate(t *testing.T) {
	location := feedbackDigestLocation("Pacific/Auckland")
	now := time.Date(2026, time.July, 19, 22, 0, 0, 0, time.UTC) // July 20 at 10:00 locally.
	sentAt := time.Date(2026, time.July, 19, 21, 0, 0, 0, time.UTC)
	subscription := feedbackDigestSubscription{
		BoardID:            uuid.New(),
		EmailFrequency:     "daily",
		CreatedAt:          now.Add(-7 * 24 * time.Hour),
		LastDigestSentAt:   &sentAt,
		LastDigestCursorAt: &sentAt,
	}

	require.Empty(t, dueFeedbackDigestSubscriptions(now, location, []feedbackDigestSubscription{subscription}))

	previousLocalDate := time.Date(2026, time.July, 18, 21, 0, 0, 0, time.UTC)
	subscription.LastDigestSentAt = &previousLocalDate
	subscription.LastDigestCursorAt = &previousLocalDate
	due := dueFeedbackDigestSubscriptions(now, location, []feedbackDigestSubscription{subscription})
	require.Len(t, due, 1)
	require.Equal(t, previousLocalDate, due[0].WindowStart)
}

func TestDueFeedbackDigestSubscriptionsCatchesWeeklyDeliveryUpAfterMonday(t *testing.T) {
	location := time.UTC
	createdAt := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	subscription := feedbackDigestSubscription{
		BoardID:        uuid.New(),
		EmailFrequency: "weekly",
		CreatedAt:      createdAt,
	}

	monday := time.Date(2026, time.July, 20, 11, 0, 0, 0, time.UTC)
	tuesday := monday.Add(24 * time.Hour)

	due := dueFeedbackDigestSubscriptions(monday, location, []feedbackDigestSubscription{subscription})
	require.Len(t, due, 1)
	require.Equal(t, createdAt, due[0].WindowStart)
	require.Len(t, dueFeedbackDigestSubscriptions(tuesday, location, []feedbackDigestSubscription{subscription}), 1)

	subscription.LastDigestSentAt = &monday
	require.Empty(t, dueFeedbackDigestSubscriptions(tuesday, location, []feedbackDigestSubscription{subscription}))

	cursorAt := monday.Add(-feedbackDigestConsistencyLag)
	subscription.LastDigestCursorAt = &cursorAt
	nextMonday := monday.Add(7 * 24 * time.Hour)
	due = dueFeedbackDigestSubscriptions(nextMonday, location, []feedbackDigestSubscription{subscription})
	require.Len(t, due, 1)
	require.Equal(t, cursorAt, due[0].WindowStart)
}

func TestDueFeedbackDigestSubscriptionsWaitsForNextPeriodWhenEnabledAfterDeliveryTime(t *testing.T) {
	location := time.UTC
	mondayAfterDelivery := time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC)
	subscription := feedbackDigestSubscription{
		BoardID:        uuid.New(),
		EmailFrequency: "weekly",
		CreatedAt:      mondayAfterDelivery,
	}

	require.Empty(t, dueFeedbackDigestSubscriptions(
		mondayAfterDelivery.Add(time.Hour),
		location,
		[]feedbackDigestSubscription{subscription},
	))
	require.Len(t, dueFeedbackDigestSubscriptions(
		mondayAfterDelivery.Add(7*24*time.Hour),
		location,
		[]feedbackDigestSubscription{subscription},
	), 1)
}

func TestFeedbackDigestSubjectPluralizesItems(t *testing.T) {
	require.Equal(t, "1 new feedback item in Acme", feedbackDigestSubject(1, "Acme"))
	require.Equal(t, "6 new feedback items in Acme", feedbackDigestSubject(6, "Acme"))
}

func TestFeedbackDigestMessageIDIsStableForDelivery(t *testing.T) {
	deliveryID := uuid.New()

	require.Equal(
		t,
		"<feedback-digest-"+deliveryID.String()+"@fortyone.app>",
		feedbackDigestMessageID(deliveryID),
	)
}

func TestFormatFeedbackDigestEmailContentUsesLinkedSafeCompactRows(t *testing.T) {
	teamID := uuid.New()
	items := make([]feedbackDigestItem, 0, 6)
	for index := range 6 {
		items = append(items, feedbackDigestItem{
			ID:                 uuid.New(),
			TeamID:             teamID,
			Title:              "Item " + string(rune('A'+index)),
			AuthorName:         "Customer & Co",
			TeamName:           "Product <Core>",
			TotalCount:         6,
			PendingReviewCount: 4,
		})
	}

	rendered := formatFeedbackDigestEmailContent(items, "https://acme.fortyone.app")

	require.Contains(t, rendered, "6 new items were submitted. 4 still need review.")
	require.Contains(t, rendered, "/teams/"+teamID.String()+"/feedback/"+items[0].ID.String())
	require.Contains(t, rendered, "Customer &amp; Co")
	require.Contains(t, rendered, "Product &lt;Core&gt;")
	require.Contains(t, rendered, "Item E")
	require.NotContains(t, rendered, "Item F")
	require.Contains(t, rendered, "+1")
	require.Contains(t, rendered, "more item is waiting in Feedback")
	require.NotContains(t, rendered, "<ul")
	require.NotContains(t, rendered, "<li")
}

func TestFeedbackDigestCopyRequestIncludesThemeContextAndTrustedDestinations(t *testing.T) {
	teamID := uuid.New()
	items := make([]feedbackDigestItem, 0, 6)
	for index := range 6 {
		items = append(items, feedbackDigestItem{
			ID:                 uuid.New(),
			TeamID:             teamID,
			Title:              "Export request " + string(rune('A'+index)),
			Description:        "<p>Customers need a faster export for finance reviews.</p>",
			AuthorName:         "Customer",
			TeamName:           "Core product",
			Status:             "pending",
			TotalCount:         6,
			PendingReviewCount: 4,
		})
	}
	recipient := feedbackDigestRecipient{
		WorkspaceName: "Acme",
	}

	request, destinations := feedbackDigestCopyRequest(recipient, items, "https://acme.fortyone.app")

	require.True(t, request.IncludeFeedbackThemeSummary)
	require.Len(t, request.Facts, 7)
	require.True(t, request.Facts[0].Required)
	require.Equal(t, []string{"Acme"}, request.Facts[0].EntityTokens)
	require.Equal(t, []string{"6 new feedback items in this digest", "4 are pending or under review"}, request.Facts[0].ProtectedTokens)
	require.Contains(t, request.Facts[1].Text, "Customers need a faster export")
	require.NotContains(t, request.Facts[1].Text, "<p>")
	require.Equal(
		t,
		[]string{"submitted by Customer", "to Core product", "currently has status pending"},
		request.Facts[1].ProtectedTokens,
	)
	for index := 1; index <= feedbackDigestItemLimit; index++ {
		require.True(t, request.Facts[index].Required)
		require.Contains(t, destinations, request.Facts[index].ReferenceID)
	}
	require.False(t, request.Facts[6].Required)
	require.NotContains(t, destinations, request.Facts[6].ReferenceID)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID.String()+"/feedback/"+items[0].ID.String(), destinations["feedback_primary"].URL)
}

func TestSendFeedbackDigestUsesGeneratedThemeCopyAndMayaSender(t *testing.T) {
	teamID := uuid.New()
	items := []feedbackDigestItem{
		{
			ID:                 uuid.New(),
			TeamID:             teamID,
			Title:              "Faster exports",
			Description:        "Finance teams need exports before weekly reviews.",
			AuthorName:         "Amara",
			TeamName:           "Core product",
			Status:             "pending",
			TotalCount:         3,
			PendingReviewCount: 2,
		},
		{
			ID:                 uuid.New(),
			TeamID:             teamID,
			Title:              "Scheduled reports",
			Description:        "Send a report to finance every Monday.",
			AuthorName:         "Theo",
			TeamName:           "Core product",
			Status:             "reviewing",
			TotalCount:         3,
			PendingReviewCount: 2,
		},
		{
			ID:                 uuid.New(),
			TeamID:             teamID,
			Title:              "Shareable dashboards",
			Description:        "Let leaders review metrics without a login.",
			AuthorName:         "Lina",
			TeamName:           "Core product",
			Status:             "planned",
			TotalCount:         3,
			PendingReviewCount: 2,
		},
	}
	generator := &recordingEmailCopyGenerator{output: emailcopy.Output{
		Subject: emailcopy.GroundedText{Text: "Customers are asking for easier reporting", ReferenceIDs: []string{"feedback_" + items[0].ID.String(), "feedback_" + items[1].ID.String()}},
		H1:      emailcopy.GroundedText{Text: "Make reporting easier to share", ReferenceIDs: []string{"feedback_" + items[0].ID.String(), "feedback_" + items[1].ID.String()}},
		Intro:   emailcopy.GroundedText{Text: "Three new notes give the team a focused place to start.", ReferenceIDs: []string{"feedback_summary"}},
		Rows: []emailcopy.Row{
			{ReferenceID: "feedback_summary", Text: "3 new feedback items arrived, and 2 are pending or under review."},
			{ReferenceID: "feedback_" + items[0].ID.String(), Text: "Faster exports points to a recurring reporting workflow."},
			{ReferenceID: "feedback_" + items[1].ID.String(), Text: "Scheduled reports would reduce manual weekly preparation."},
			{ReferenceID: "feedback_" + items[2].ID.String(), Text: "Shareable dashboards extends the same need to leadership."},
		},
		CTAs: []emailcopy.CTA{{ReferenceID: "feedback_primary", Label: "Review the signal"}},
		FeedbackThemeSummary: &emailcopy.GroundedText{
			Text:         "The clearest theme is making recurring reporting easier to prepare and share.",
			ReferenceIDs: []string{"feedback_" + items[0].ID.String(), "feedback_" + items[1].ID.String(), "feedback_" + items[2].ID.String()},
		},
	}}
	mailerService := &recordingMailer{}
	recipient := feedbackDigestRecipient{
		UserEmail:     "joseph@example.com",
		UserName:      "Joseph",
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
	}

	err := sendFeedbackDigestEmail(context.Background(), newTestJobLogger(), mailerService, generator, uuid.New(), recipient, items)

	require.NoError(t, err)
	require.Len(t, generator.requests, 1)
	require.True(t, generator.requests[0].IncludeFeedbackThemeSummary)
	require.Len(t, mailerService.templatedEmails, 1)
	email := mailerService.templatedEmails[0]
	require.Equal(t, mailer.SenderProfileMaya, email.Sender)
	require.Equal(t, "Customers are asking for easier reporting", email.Subject)
	data := requireEmailData(t, email)
	require.Contains(t, data["NotificationMessage"], "clearest theme")
	require.Contains(t, data["NotificationMessage"], ">Faster exports</a>")
	require.Equal(t, "Review the signal", data["NotificationCTALabel"])
}

func TestFeedbackDigestQueriesKeepWorkspaceAndTeamAccessBoundaries(t *testing.T) {
	recipientsQuery := feedbackDigestRecipientsQuery()
	require.Contains(t, recipientsQuery, "tm.user_id = fbs.user_id")
	require.Contains(t, recipientsQuery, "wm.workspace_id = fb.workspace_id")
	require.Contains(t, recipientsQuery, "wm.role IN ('admin', 'member')")
	require.Contains(t, recipientsQuery, "u.is_active = true")
	require.Contains(t, recipientsQuery, "u.is_system = false")

	subscriptionsQuery := feedbackDigestSubscriptionsQuery()
	require.Contains(t, subscriptionsQuery, "fb.workspace_id = $2")
	require.Contains(t, subscriptionsQuery, "tm.user_id = fbs.user_id")
	require.Contains(t, subscriptionsQuery, "wm.role IN ('admin', 'member')")

	itemsQuery := feedbackDigestItemsQuery()
	require.Contains(t, itemsQuery, "fi.workspace_id = $2")
	require.Contains(t, itemsQuery, "fb.workspace_id = fi.workspace_id")
	require.Contains(t, itemsQuery, "fbs.user_id = $1")
	require.Contains(t, itemsQuery, "fi.created_at > bw.window_start")
	require.Contains(t, itemsQuery, "fi.created_at <= $5")
	require.Contains(t, itemsQuery, "fi.submission_source IN ('portal', 'widget', 'integration')")
	require.Contains(t, itemsQuery, "fi.description")
	require.False(t, strings.Contains(itemsQuery, "submission_source = 'internal'"))
}

func TestFeedbackDigestClaimRetriesOnlyFailedOrStaleProcessingDeliveries(t *testing.T) {
	query := feedbackDigestClaimQuery()
	require.Contains(t, query, "ON CONFLICT (workspace_id, recipient_id, local_date)")
	require.Contains(t, query, "feedback_digest_deliveries.status = 'failed'")
	require.Contains(t, query, "feedback_digest_deliveries.status = 'processing'")
	require.Contains(t, query, "updated_at < NOW() - INTERVAL '7200 seconds'")
	require.NotContains(t, query, "status = 'sent'")
	require.NotContains(t, query, "status = 'skipped'")
}
