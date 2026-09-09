package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func strategyCheckInTestInput(t *testing.T) notificationDigestCopyInput {
	t.Helper()
	health, zero, one := "On Track", 0.0, 1.0
	objectiveID := uuid.New()
	weekly := &strategyWeeklyCheckInSnapshot{
		StaleAfterDays: 7,
		Counts:         strategyWeeklyCheckInCounts{StaleObjectives: 1, StaleKeyResults: 3, UniqueObjectives: 1},
		Objectives: []strategyObjectiveSnapshot{{
			ID: objectiveID, Name: "Launch the customer portal", Health: &health,
			Reasons: []string{"stale"}, UpdatedAt: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC),
		}},
	}
	for _, name := range []string{"Complete the onboarding flow", "Reduce page load time to under 2 seconds (P95)", "Publish the help center"} {
		weekly.KeyResults = append(weekly.KeyResults, strategyKeyResultSnapshot{
			ID: uuid.New(), ObjectiveID: objectiveID, ObjectiveName: "Launch the customer portal",
			Name: name, MeasurementType: "boolean", CurrentValue: &zero, TargetValue: &one,
			ObjectiveHealth: &health, UpdatedAt: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
		})
	}
	message, err := json.Marshal(NotificationMessage{Strategy: &strategyNotificationSnapshot{
		Version: 1, Kind: "weekly_check_in", WeeklyCheckIn: weekly,
	}})
	require.NoError(t, err)
	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		RecipientID: uuid.New(), WorkspaceID: uuid.New(), WorkspaceName: "Product",
		Items: []NotificationEmailDigestItem{{NotificationID: uuid.New(), EntityType: "strategy", Message: message}},
	}, "https://product.fortyone.app")
	require.NoError(t, err)
	return input
}

func TestStrategyCheckInIsConciseInGeneratedAndFallbackCopy(t *testing.T) {
	input := strategyCheckInTestInput(t)
	require.False(t, input.Request.IncludeSenderProse)
	require.True(t, input.Request.IncludeReplyPrompt)
	require.Contains(t, input.Request.ProductVoice, "do not repeat objective or key-result names")
	fallback := templateDigest(input.Fallback)
	require.Len(t, fallback.Rows, 4)
	require.Equal(t, 1, strings.Count(renderNotificationDigestPlainText(input.Fallback), "7 days"))
	require.NotContains(t, renderNotificationDigestPlainText(input.Fallback), "recent means")
	require.NotContains(t, input.Fallback.Intro, "I’m Maya")
	generated := emailcopy.Output{
		Subject:     emailcopy.GroundedText{Text: "Your strategy check-in"},
		H1:          emailcopy.GroundedText{Text: "A quick progress check"},
		Intro:       emailcopy.GroundedText{Text: "Your portal launch needs a progress update."},
		ReplyPrompt: &emailcopy.GroundedText{Text: "Have an update or a blocker? Reply to this email and I’ll help you update your strategy."},
	}
	for _, fact := range input.Request.Facts {
		if fact.Required {
			generated.Rows = append(generated.Rows, emailcopy.Row{ReferenceID: fact.ReferenceID, Text: fact.Text})
		}
	}
	copy, err := buildGeneratedNotificationDigestCopy(input, generated)
	require.NoError(t, err)
	require.Equal(t, "Your strategy check-in", copy.Heading)
	for _, row := range copy.Rows {
		require.Equal(t, 1, strings.Count(renderNotificationDigestPlainText(copy), row.Label))
		require.Contains(t, renderNotificationDigestCopy(copy), row.Label)
	}
	for _, digest := range []mailer.Digest{fallback, templateDigest(copy)} {
		for _, row := range digest.Rows {
			require.NotContains(t, row.Text, row.Label)
			require.Equal(t, "status", row.Icon)
			require.NotContains(t, row.Text, "7 days")
			require.Less(t, len([]rune(row.Text)), 180)
		}
		for _, row := range digest.Rows[1:] {
			require.Contains(t, row.Text, "Not complete")
			require.NotContains(t, row.Text, "boolean")
			require.NotContains(t, row.Text, "On Track")
		}
	}
	require.Equal(t, "priority", strategyObjectiveIcon(strategyObjectiveSnapshot{Reasons: []string{"at_risk"}}))
}

func TestBuildNotificationDigestCopyInputSummarizesOmittedMixedDigestItems(t *testing.T) {
	strategyNotificationID := uuid.New()
	objectives := make([]strategyObjectiveSnapshot, maxNotificationDigestRows)
	for index := range objectives {
		objectives[index] = strategyObjectiveSnapshot{
			ID:        uuid.New(),
			TeamID:    uuid.New(),
			Name:      "Objective " + string(rune('A'+index)),
			UpdatedAt: time.Date(2026, time.August, 1, 8, 0, 0, 0, time.UTC),
			Reasons:   []string{"stale"},
		}
	}
	strategyMessage, err := json.Marshal(NotificationMessage{
		Template: "Weekly strategy check-in",
		Strategy: &strategyNotificationSnapshot{
			Version: 1,
			Kind:    "weekly_check_in",
			WeeklyCheckIn: &strategyWeeklyCheckInSnapshot{
				StaleAfterDays: 7,
				Counts: strategyWeeklyCheckInCounts{
					StaleObjectives:  len(objectives),
					UniqueObjectives: len(objectives),
				},
				Objectives: objectives,
			},
		},
	})
	require.NoError(t, err)
	activityMessage, err := json.Marshal(NotificationMessage{
		Template: "{actor} moved the task to {value}",
		Variables: map[string]Variable{
			"actor": {Value: "Maya Chen", Type: "actor"},
			"value": {Value: "In review", Type: "value"},
		},
	})
	require.NoError(t, err)

	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		WorkspaceName: "Product",
		Items: []NotificationEmailDigestItem{
			{
				NotificationID:   strategyNotificationID,
				NotificationType: "strategy_update",
				EntityType:       "strategy",
				Title:            "Your weekly strategy check-in",
				Message:          strategyMessage,
			},
			{
				NotificationID:   uuid.New(),
				NotificationType: "story_update",
				EntityType:       "story",
				EntityID:         uuid.New(),
				Title:            "Ship billing states",
				Message:          activityMessage,
			},
		},
	}, "https://product.fortyone.app")

	require.NoError(t, err)
	require.True(t, input.HasStrategySnapshot)
	require.Len(t, input.Fallback.Rows, maxNotificationDigestRows)
	require.Contains(t, factText(input.Request.Facts, "remaining_updates"), "1 additional unread update")
	require.Equal(t, "https://product.fortyone.app/notifications", input.Actions[input.FactActions["remaining_updates"]])
	require.Equal(t, "Notifications", input.FactLabels["remaining_updates"])
}

func TestBuildNotificationDigestCopyInputUsesStructuredPlanningFacts(t *testing.T) {
	notificationID := uuid.New()
	rawMessage, err := json.Marshal(NotificationMessage{
		Template: "Your next planning period starts soon.",
		Strategy: &strategyNotificationSnapshot{
			Version:     1,
			Kind:        "planning_reminder",
			GeneratedAt: time.Date(2026, time.September, 10, 9, 0, 0, 0, time.UTC),
			Planning: &strategyPlanningSnapshot{
				Period:          "Q4",
				StartsAt:        time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
				DaysUntil:       21,
				HasUltimateGoal: true,
				PillarCount:     3,
				ObjectiveCount:  0,
				MissingElements: []string{"objectives"},
			},
		},
	})
	require.NoError(t, err)

	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		WorkspaceName: "Product",
		Items: []NotificationEmailDigestItem{{
			NotificationID:   notificationID,
			NotificationType: "strategy_update",
			EntityType:       "strategy",
			Title:            "Plan your Q4",
			Message:          rawMessage,
		}},
	}, "https://product.fortyone.app")

	require.NoError(t, err)
	require.Len(t, input.Request.Facts, 2)
	planningFact := input.Request.Facts[1]
	require.Contains(t, planningFact.Text, "Q4 starts on October 1, 2026, in 21 days")
	require.Contains(t, planningFact.Text, "has 3 strategic pillars")
	require.Contains(t, planningFact.Text, "missing elements are objectives")
	require.NotContains(t, planningFact.Text, "Your next planning period starts soon")
	require.Equal(t, "https://product.fortyone.app/strategy", input.Actions[input.FactActions[planningFact.ReferenceID]])
}

func TestBuildNotificationDigestCopyInputDistinguishesMonthlySnapshotFromPeriodActivity(t *testing.T) {
	notificationID := uuid.New()
	keyResultProgress := 46.0
	rawMessage, err := json.Marshal(NotificationMessage{
		Template: "Here is the strategy-to-execution picture for last month.",
		Strategy: &strategyNotificationSnapshot{
			Version:     1,
			Kind:        "monthly_summary",
			GeneratedAt: time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC),
			MonthlySummary: &strategyMonthlySummarySnapshot{
				PeriodStart:          time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:            time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
				PillarCount:          4,
				PillarsNeedingReview: 1,
				ObjectiveCount:       8,
				AtRiskObjectives:     2,
				UnalignedObjectives:  3,
				KeyResultCount:       6,
				KeyResultProgress:    &keyResultProgress,
				CompletedStories:     12,
			},
		},
	})
	require.NoError(t, err)

	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		WorkspaceName: "Product",
		Items: []NotificationEmailDigestItem{{
			NotificationID:   notificationID,
			NotificationType: "strategy_update",
			EntityType:       "strategy",
			Title:            "July strategy summary",
			Message:          rawMessage,
		}},
	}, "https://product.fortyone.app")

	require.NoError(t, err)
	require.Len(t, input.Request.Facts, 2)
	monthlyFact := input.Request.Facts[1]
	require.Contains(t, monthlyFact.Text, "current snapshot has 4 strategic pillars")
	require.Contains(t, monthlyFact.Text, "46% average key-result progress across 6 key results")
	require.Contains(t, monthlyFact.Text, "From July 1, 2026 up to August 1, 2026, 12 linked tasks were completed")
	require.NotContains(t, monthlyFact.Text, "picture for last month")
}

func TestBuildNotificationDigestCopyInputDoesNotReportZeroProgressWithoutKeyResults(t *testing.T) {
	notificationID := uuid.New()
	rawMessage, err := json.Marshal(NotificationMessage{
		Template: "Here is the current strategy snapshot and last month's linked delivery.",
		Strategy: &strategyNotificationSnapshot{
			Version:     1,
			Kind:        "monthly_summary",
			GeneratedAt: time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC),
			MonthlySummary: &strategyMonthlySummarySnapshot{
				PeriodStart:      time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:        time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
				PillarCount:      4,
				ObjectiveCount:   8,
				CompletedStories: 12,
			},
		},
	})
	require.NoError(t, err)

	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		WorkspaceName: "Product",
		Items: []NotificationEmailDigestItem{{
			NotificationID:   notificationID,
			NotificationType: "strategy_update",
			EntityType:       "strategy",
			Title:            "July strategy summary",
			Message:          rawMessage,
		}},
	}, "https://product.fortyone.app")

	require.NoError(t, err)
	require.Len(t, input.Request.Facts, 2)
	monthlyFact := input.Request.Facts[1]
	require.Contains(t, monthlyFact.Text, "there are no key results in the current snapshot")
	require.NotContains(t, monthlyFact.Text, "0% average key-result progress")
}

func TestBuildGeneratedNotificationDigestCopyResolvesOnlyTrustedReferences(t *testing.T) {
	input := notificationDigestCopyInput{
		Actions: map[string]string{
			"activity_action":   "https://product.fortyone.app/notifications/notification-1",
			digestActionPrimary: "https://product.fortyone.app/notifications",
		},
		FactActions: map[string]string{"activity": "activity_action"},
		FactLabels:  map[string]string{"activity": "Ship billing states"},
		Request: emailcopy.Request{Facts: []emailcopy.Fact{{
			ReferenceID: "activity",
			Required:    true,
		}}},
		Fallback: notificationDigestCopy{
			CTA: notificationDigestCopyCTA{Label: "Open notifications", URL: "https://product.fortyone.app/notifications"},
		},
		NotificationsURL: "https://product.fortyone.app/notifications",
	}
	generated := emailcopy.Output{
		Subject: emailcopy.GroundedText{Text: "One update is ready"},
		H1:      emailcopy.GroundedText{Text: "A task moved forward"},
		Intro:   emailcopy.GroundedText{Text: "Here’s the useful part."},
		Rows:    []emailcopy.Row{{ReferenceID: "activity", Text: "Ship billing states moved to In review."}},
		CTAs:    []emailcopy.CTA{{ReferenceID: digestActionPrimary, Label: "Open notifications"}},
	}

	copy, err := buildGeneratedNotificationDigestCopy(input, generated)

	require.NoError(t, err)
	require.Equal(t, "https://product.fortyone.app/notifications/notification-1", copy.Rows[0].URL)
	require.Equal(t, "https://product.fortyone.app/notifications", copy.CTA.URL)

	generated.Rows[0].ReferenceID = "invented_fact"
	_, err = buildGeneratedNotificationDigestCopy(input, generated)
	require.ErrorContains(t, err, "unknown fact")
}

func factText(facts []emailcopy.Fact, referenceID string) string {
	for _, fact := range facts {
		if fact.ReferenceID == referenceID {
			return fact.Text
		}
	}
	return ""
}

func TestGeneratedNotificationDigestCopyUsesMayaOrFallsBackDeterministically(t *testing.T) {
	input := notificationDigestCopyInput{
		Actions: map[string]string{
			"strategy_action":    "https://product.fortyone.app/notifications/notification-1",
			digestActionStrategy: "https://product.fortyone.app/strategy",
		},
		FactActions: map[string]string{"strategy_fact": "strategy_action"},
		FactLabels:  map[string]string{"strategy_fact": "Grow enterprise revenue"},
		Request: emailcopy.Request{Facts: []emailcopy.Fact{{
			ReferenceID: "strategy_fact",
			Required:    true,
		}}},
		HasStrategySnapshot: true,
		Fallback: notificationDigestCopy{
			Subject: "Your strategy check-in",
			Heading: "Your strategy check-in",
			Intro:   "Here are the objectives, key results, and strategy updates that need your attention. I’m Maya, your AI agent. Reply to this email with what changed or what you want updated.",
			Rows:    []notificationDigestCopyRow{{Text: "Grow enterprise revenue is At Risk.", URL: "https://product.fortyone.app/notifications/notification-1"}},
			CTA:     notificationDigestCopyCTA{Label: "Review strategy", URL: "https://product.fortyone.app/strategy"},
			Sender:  mailer.SenderProfileMaya,
		},
	}
	generated := emailcopy.Output{
		Subject:     emailcopy.GroundedText{Text: "Revenue needs a clear next move"},
		H1:          emailcopy.GroundedText{Text: "Bring revenue back into focus"},
		Intro:       emailcopy.GroundedText{Text: "One objective needs a review."},
		SenderProse: &emailcopy.GroundedText{Text: "I’ve gathered the signals that matter most."},
		Rows:        []emailcopy.Row{{ReferenceID: "strategy_fact", Text: "Grow enterprise revenue is At Risk.", CTAReferenceID: "strategy_action"}},
		CTAs:        []emailcopy.CTA{{ReferenceID: digestActionStrategy, Label: "Review strategy"}},
	}

	copy, err := buildGeneratedNotificationDigestCopy(input, generated)

	require.NoError(t, err)
	require.Equal(t, mailer.SenderProfileMaya, copy.Sender)
	require.Contains(t, copy.Intro, "I’ve gathered")
	require.NotContains(t, copy.Intro, "Reply")

	stub := &notificationEmailCopyStub{err: errors.New("provider unavailable")}
	fallback, generateErr := generateNotificationDigestCopy(context.Background(), stub, input)
	require.ErrorContains(t, generateErr, "provider unavailable")
	require.Equal(t, input.Request, stub.request)
	require.Equal(t, "Your strategy check-in", fallback.Subject)
	require.Equal(t, "Review strategy", fallback.CTA.Label)
	require.Equal(t, mailer.SenderProfileMaya, fallback.Sender)
}

func TestRenderNotificationDigestCopyLinksOnlyTheCanonicalEntityLabel(t *testing.T) {
	rendered := renderNotificationDigestCopy(notificationDigestCopy{
		Intro: "One objective needs a useful next move.",
		Rows: []notificationDigestCopyRow{{
			Text:  "Grow enterprise revenue is At Risk and has not had a recent update.",
			Label: "Grow enterprise revenue",
			URL:   "https://product.fortyone.app/objectives/revenue",
		}},
	})

	require.Contains(t, rendered, `href="https://product.fortyone.app/objectives/revenue"`)
	require.Contains(t, rendered, `>Grow enterprise revenue</a> is At Risk`)
	require.NotContains(t, rendered, `>Grow enterprise revenue is At Risk`)
}

func TestSelectWeeklyStrategyDetailsBalancesObjectivesAndKeyResults(t *testing.T) {
	objectives := make([]strategyObjectiveSnapshot, 20)
	keyResults := make([]strategyKeyResultSnapshot, 20)

	selectedObjectives, selectedKeyResults := selectWeeklyStrategyDetails(objectives, keyResults, 10)

	require.Len(t, selectedObjectives, 5)
	require.Len(t, selectedKeyResults, 5)
}

func TestTaskDigestKeepsLatestEventBeforeApplyingDetailLimit(t *testing.T) {
	now := time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC)
	taskID := uuid.New()
	items := make([]NotificationEmailDigestItem, 0, 10)
	for i := range 7 {
		items = append(items, NotificationEmailDigestItem{
			NotificationID: uuid.New(), EntityType: "story", EntityID: taskID,
			Title: "Ticketing system mobile app", CreatedAt: now.Add(time.Duration(i) * time.Minute),
			Message: json.RawMessage(`{"template":"hector changed priority to High"}`),
		})
	}
	latest := &items[6]
	latest.Message = json.RawMessage(`{"template":"{actor} changed the start date","variables":{"actor":{"type":"actor","value":"hector"}}}`)
	latest.NotificationType, latest.ActorName = "story_update", "hector"
	latestID := latest.NotificationID
	// A different task can have exactly the same title.
	items = append(items, NotificationEmailDigestItem{
		NotificationID: uuid.New(), EntityType: "story", EntityID: uuid.New(),
		Title: "Ticketing system mobile app", CreatedAt: now,
		Message: json.RawMessage(`{"template":"Another task was assigned to you"}`),
	})
	// Input order must not decide which event is newest.
	items[0], items[6] = items[6], items[0]
	original := append([]NotificationEmailDigestItem(nil), items...)
	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{WorkspaceName: "Art Circles", Items: items}, "https://art.fortyone.app")
	require.NoError(t, err)
	require.Equal(t, original, items, "presentation must not mutate delivery coverage")
	require.Len(t, input.Fallback.Rows, 2)
	require.Equal(t, "2 tasks updated in Art Circles", input.Fallback.Subject)
	require.Contains(t, input.Request.Facts[0].Text, "2 unread product updates")
	require.Len(t, input.Request.Facts, 3)
	row := input.Fallback.Rows[0]
	require.Contains(t, row.Text, "changed the start date")
	require.Contains(t, row.URL, latestID.String())
	require.Equal(t, "calendar", row.Icon)
	require.Equal(t, "hector", row.Actor.Name)
	for _, fact := range input.Request.Facts {
		require.NotContains(t, fact.Text, "High")
	}
	generated, err := buildGeneratedNotificationDigestCopy(input, emailcopy.Output{
		Subject: emailcopy.GroundedText{Text: "Task updates"}, H1: emailcopy.GroundedText{Text: "Task updates"},
		Intro: emailcopy.GroundedText{Text: "Here are the latest task updates."},
		Rows:  []emailcopy.Row{{ReferenceID: "notification_1", Text: row.Text}},
	})
	require.NoError(t, err)
	require.Equal(t, row.Icon, templateDigest(generated).Rows[0].Icon)
	require.Equal(t, row.Actor, templateDigest(generated).Rows[0].Actor)
}

func TestLatestTaskDigestItemsPreservesOtherEntitiesAndUsesStableTies(t *testing.T) {
	entityID := uuid.New()
	older := NotificationEmailDigestItem{EntityType: "story", EntityID: entityID, NotificationID: uuid.MustParse("00000000-0000-0000-0000-000000000001")}
	newer := older
	newer.NotificationID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	feedback := older
	feedback.EntityType = "feedback"
	unknown := older
	unknown.EntityID = uuid.Nil
	items := []NotificationEmailDigestItem{newer, older, feedback, feedback, unknown, unknown}
	require.Equal(t, []NotificationEmailDigestItem{newer, feedback, feedback, unknown, unknown}, latestTaskDigestItems(items))
}

func TestNotificationIconsMatchPersistedEvents(t *testing.T) {
	for _, tc := range []struct {
		name, kind, message, want string
	}{
		{"start date", "story_update", `{"template":"{actor} changed the start date"}`, "calendar"},
		{"deadline removed", "story_update", `{"variables":{"field":{"type":"field","value":"deadline"}}}`, "calendar"},
		{"deadline set", "story_update", `{"variables":{"value":{"type":"date","value":"8 Sep"}}}`, "calendar"},
		{"comment", "story_comment", `{"template":"{actor} commented: {content}"}`, "comment"},
		{"reply", "comment_reply", `{}`, "comment"},
		{"feedback comment", "feedback_comment", `{}`, "comment"},
		{"comment mentioning date", "story_comment", `{"variables":{"content":{"type":"value","value":"Change the start date"}}}`, "comment"},
		{"priority", "story_update", `{"template":"{actor} changed priority to {value}"}`, "priority"},
		{"status field fallback", "story_update", `{"template":"{actor} updated {field}","variables":{"field":{"type":"field","value":"Status"}}}`, "status"},
		{"priority field", "story_update", `{"variables":{"field":{"type":"field","value":"Priority"}}}`, "priority"},
		{"comment mentioning priority", "story_comment", `{"variables":{"content":{"type":"value","value":"Please change priority and status"}}}`, "comment"},
		{"title mentioning status", "story_update", `{"template":"{actor} renamed the story to {value}","variables":{"value":{"type":"value","value":"Fix priority status"}}}`, ""},
		{"assignment", "story_update", `{"template":"{actor} assigned you a task"}`, ""},
		{"status", "story_update", `{"template":"{actor} moved the task to {value}"}`, "status"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var message NotificationMessage
			require.NoError(t, json.Unmarshal([]byte(tc.message), &message))
			require.Equal(t, tc.want, notificationIcon(message, tc.kind))
		})
	}
}

func TestGeneratedActivityPreservesChangeValuesAndSeparatesReason(t *testing.T) {
	message := NotificationMessage{Template: "{actor} moved this task to {scheduled_for}: {reason}", Variables: map[string]Variable{
		"actor":         {Value: "Maya", Type: "actor"},
		"scheduled_for": {Value: "8 Sep 2026 at 14:40–15:40 (UTC+02:00)", Type: "date"},
		"reason":        {Value: "The assignee's availability or this story's scheduling constraints changed, so Maya moved it to the next safe slot.", Type: "value"},
	}}
	raw, err := json.Marshal(message)
	require.NoError(t, err)
	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{WorkspaceName: "Art Circles", Items: []NotificationEmailDigestItem{{
		NotificationID: uuid.New(), EntityID: uuid.New(), EntityType: "story", Title: "Scraping Segments Updates", Message: raw,
	}}}, "https://art.fortyone.app")
	require.NoError(t, err)
	generated, err := buildGeneratedNotificationDigestCopy(input, emailcopy.Output{
		Subject: emailcopy.GroundedText{Text: "Task updated"}, H1: emailcopy.GroundedText{Text: "Task updated"}, Intro: emailcopy.GroundedText{Text: "One update."},
		Rows: []emailcopy.Row{{ReferenceID: "notification_1", Text: "Scraping Segments Updates — moved to the next safe slot."}},
	})
	require.NoError(t, err)
	require.Equal(t, input.Fallback.Rows, generated.Rows)
	digest := templateDigest(generated)
	require.Equal(t, "Maya moved this task to 8 Sep 2026 at 14:40–15:40 (UTC+02:00)", digest.Rows[0].Text)
	require.Equal(t, "The assignee's availability or the task's scheduling constraints changed.", digest.Rows[0].Detail)
	require.Equal(t, []string{"8 Sep 2026 at 14:40–15:40 (UTC+02:00)"}, digest.Rows[0].Highlights)
	require.NotContains(t, renderNotificationDigestPlainText(generated), "safe slot")
	require.Contains(t, renderNotificationDigestPlainText(generated), "\nThe assignee's availability")
}

func TestNotificationChangeHighlightsExcludeCommentAndReason(t *testing.T) {
	message := NotificationMessage{Template: "{actor} moved the task from {previous_value} to {value}: {reason}", Variables: map[string]Variable{
		"actor": {Value: "hector", Type: "actor"}, "previous_value": {Value: "Backlog", Type: "value"}, "value": {Value: "To Do", Type: "value"},
		"reason": {Value: "A plan is ready.", Type: "value"}, "content": {Value: "Don't bold a whole comment", Type: "value"},
	}}
	text, detail, highlights := notificationActivityCopy(message)
	require.Equal(t, "hector moved the task from Backlog to To Do", text)
	require.Equal(t, "A plan is ready.", detail)
	require.ElementsMatch(t, []string{"Backlog", "To Do"}, highlights)
	require.Equal(t, "status", notificationIcon(message, "story_update"))
}

func TestNotificationVariablesAreNotRecursivelyExpanded(t *testing.T) {
	message := NotificationMessage{Template: "{actor} commented: {content}", Variables: map[string]Variable{
		"actor": {Value: "Maya", Type: "actor"}, "content": {Value: "Keep {actor} as literal text", Type: "value"},
	}}
	require.Equal(t, "Maya commented: Keep {actor} as literal text", parseNotificationMessage(message).Text)
}
