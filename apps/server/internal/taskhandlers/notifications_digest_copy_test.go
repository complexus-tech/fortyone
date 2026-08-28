package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
