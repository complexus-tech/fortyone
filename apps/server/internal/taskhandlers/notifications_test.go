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

type notificationEmailCopyStub struct {
	output  emailcopy.Output
	err     error
	request emailcopy.Request
}

func (stub *notificationEmailCopyStub) Generate(_ context.Context, request emailcopy.Request) (emailcopy.Output, error) {
	stub.request = request
	return stub.output, stub.err
}

func TestBuildNotificationDigestSubject(t *testing.T) {
	require.Equal(t, "New activity in Product", buildNotificationDigestSubject("Product", 1))
	require.Equal(t, "3 updates to review in Product", buildNotificationDigestSubject("Product", 3))
}

func TestNotificationEmailQueriesRequireCurrentEntityAccess(t *testing.T) {
	for _, query := range []string{notificationEmailDataQuery(), notificationEmailDigestDataQuery()} {
		require.Contains(t, query, "LEFT JOIN workspace_members wm")
		require.Contains(t, query, "CAST(n.entity_type AS TEXT) = 'feedback'")
		require.Contains(t, query, "wm.role IN ('admin', 'member', 'guest')")
		require.Contains(t, query, "story_member.user_id = n.recipient_id")
		require.Contains(t, query, "objective_member.user_id = n.recipient_id")
		require.Contains(t, query, "key_result_member.user_id = n.recipient_id")
		require.Contains(t, query, "w.deleted_at IS NULL")
	}
}

func TestFilterWeeklyStrategyItemForCurrentTeams(t *testing.T) {
	allowedTeamID := uuid.New()
	removedTeamID := uuid.New()
	allowedObjectiveID := uuid.New()
	removedObjectiveID := uuid.New()
	message, err := json.Marshal(NotificationMessage{
		Strategy: &strategyNotificationSnapshot{
			Version: 1,
			Kind:    "weekly_check_in",
			WeeklyCheckIn: &strategyWeeklyCheckInSnapshot{
				StaleAfterDays: 7,
				Counts: strategyWeeklyCheckInCounts{
					AtRiskObjectives: 2,
					StaleObjectives:  1,
					StaleKeyResults:  2,
					UniqueObjectives: 2,
				},
				TeamCounts: []strategyWeeklyCheckInTeamCountsSnapshot{
					{TeamID: allowedTeamID, Counts: strategyWeeklyCheckInCounts{AtRiskObjectives: 1, StaleObjectives: 1, StaleKeyResults: 1, UniqueObjectives: 1}, OmittedDetails: &strategyWeeklyCheckInOmittedDetailsSnapshot{Objectives: 2, KeyResults: 3}},
					{TeamID: removedTeamID, Counts: strategyWeeklyCheckInCounts{AtRiskObjectives: 1, StaleKeyResults: 1, UniqueObjectives: 1}, OmittedDetails: &strategyWeeklyCheckInOmittedDetailsSnapshot{Objectives: 5, KeyResults: 7}},
				},
				Objectives: []strategyObjectiveSnapshot{
					{ID: allowedObjectiveID, TeamID: allowedTeamID, Name: "Improve onboarding", Reasons: []string{"at_risk", "stale"}},
					{ID: removedObjectiveID, TeamID: removedTeamID, Name: "Expand enterprise", Reasons: []string{"at_risk"}},
				},
				KeyResults: []strategyKeyResultSnapshot{
					{ID: uuid.New(), ObjectiveID: allowedObjectiveID, TeamID: allowedTeamID, Name: "Reach 40% activation", Reasons: []string{"stale", "incomplete"}},
					{ID: uuid.New(), ObjectiveID: removedObjectiveID, TeamID: removedTeamID, Name: "Reach 25 customers", Reasons: []string{"stale", "incomplete"}},
				},
			},
		},
	})
	require.NoError(t, err)

	filtered, include, err := filterWeeklyStrategyItemForTeams(
		NotificationEmailDigestItem{NotificationID: uuid.New(), EntityType: "strategy", Message: message},
		map[uuid.UUID]struct{}{allowedTeamID: {}},
	)
	require.NoError(t, err)
	require.True(t, include)

	var filteredMessage NotificationMessage
	require.NoError(t, json.Unmarshal(filtered.Message, &filteredMessage))
	weekly := filteredMessage.Strategy.WeeklyCheckIn
	require.Len(t, weekly.Objectives, 1)
	require.Equal(t, "Improve onboarding", weekly.Objectives[0].Name)
	require.Len(t, weekly.KeyResults, 1)
	require.Equal(t, "Reach 40% activation", weekly.KeyResults[0].Name)
	require.Equal(t, strategyWeeklyCheckInCounts{
		AtRiskObjectives: 1,
		StaleObjectives:  1,
		StaleKeyResults:  1,
		UniqueObjectives: 1,
	}, weekly.Counts)
	require.Len(t, weekly.TeamCounts, 1)
	require.Equal(t, allowedTeamID, weekly.TeamCounts[0].TeamID)
	require.Equal(t, &strategyWeeklyCheckInOmittedDetailsSnapshot{Objectives: 2, KeyResults: 3}, weekly.OmittedDetails)
}

func TestFilterWeeklyStrategyItemPreservesFullAllowedCountsBeyondCappedDetails(t *testing.T) {
	teamID := uuid.New()
	message, err := json.Marshal(NotificationMessage{
		Strategy: &strategyNotificationSnapshot{
			Version: 1,
			Kind:    "weekly_check_in",
			WeeklyCheckIn: &strategyWeeklyCheckInSnapshot{
				StaleAfterDays: 7,
				Counts:         strategyWeeklyCheckInCounts{AtRiskObjectives: 100, StaleObjectives: 80, StaleKeyResults: 60, UniqueObjectives: 120},
				TeamCounts: []strategyWeeklyCheckInTeamCountsSnapshot{{
					TeamID:         teamID,
					Counts:         strategyWeeklyCheckInCounts{AtRiskObjectives: 100, StaleObjectives: 80, StaleKeyResults: 60, UniqueObjectives: 120},
					OmittedDetails: &strategyWeeklyCheckInOmittedDetailsSnapshot{Objectives: 99, KeyResults: 59},
				}},
				Objectives:     []strategyObjectiveSnapshot{{ID: uuid.New(), TeamID: teamID, Name: "Prioritized objective", Reasons: []string{"at_risk", "stale"}}},
				KeyResults:     []strategyKeyResultSnapshot{{ID: uuid.New(), ObjectiveID: uuid.New(), TeamID: teamID, Name: "Prioritized key result", Reasons: []string{"stale", "incomplete"}}},
				OmittedDetails: &strategyWeeklyCheckInOmittedDetailsSnapshot{Objectives: 99, KeyResults: 59},
			},
		},
	})
	require.NoError(t, err)

	filtered, include, err := filterWeeklyStrategyItemForTeams(
		NotificationEmailDigestItem{NotificationID: uuid.New(), EntityType: "strategy", Message: message},
		map[uuid.UUID]struct{}{teamID: {}},
	)
	require.NoError(t, err)
	require.True(t, include)
	var filteredMessage NotificationMessage
	require.NoError(t, json.Unmarshal(filtered.Message, &filteredMessage))
	require.Equal(t, strategyWeeklyCheckInCounts{AtRiskObjectives: 100, StaleObjectives: 80, StaleKeyResults: 60, UniqueObjectives: 120}, filteredMessage.Strategy.WeeklyCheckIn.Counts)
	require.Equal(t, &strategyWeeklyCheckInOmittedDetailsSnapshot{Objectives: 99, KeyResults: 59}, filteredMessage.Strategy.WeeklyCheckIn.OmittedDetails)
}

func TestFilterWeeklyStrategyItemSuppressesInaccessibleSnapshot(t *testing.T) {
	message, err := json.Marshal(NotificationMessage{
		Strategy: &strategyNotificationSnapshot{
			Version: 1,
			Kind:    "weekly_check_in",
			WeeklyCheckIn: &strategyWeeklyCheckInSnapshot{
				StaleAfterDays: 7,
				Objectives: []strategyObjectiveSnapshot{{
					ID: uuid.New(), TeamID: uuid.New(), Name: "Private objective", Reasons: []string{"at_risk"},
				}},
				KeyResults: []strategyKeyResultSnapshot{},
			},
		},
	})
	require.NoError(t, err)

	_, include, err := filterWeeklyStrategyItemForTeams(
		NotificationEmailDigestItem{NotificationID: uuid.New(), EntityType: "strategy", Message: message},
		map[uuid.UUID]struct{}{},
	)
	require.NoError(t, err)
	require.False(t, include)
}

func TestNotificationEmailDestinationLinksStrategyThroughNotification(t *testing.T) {
	workspaceURL := "https://product.fortyone.app"
	workspaceID := uuid.New()
	notificationID := uuid.New()

	destination, label := notificationEmailDestination(
		"strategy",
		workspaceID,
		"",
		notificationID,
		workspaceURL,
	)

	require.Contains(t, destination, workspaceURL+"/notifications/"+notificationID.String())
	require.Contains(t, destination, "entityId="+workspaceID.String())
	require.Contains(t, destination, "entityType=strategy")
	require.Equal(t, "strategy", label)
}

func TestBuildNotificationDigestCopyInputLinksFeedbackToPublicItem(t *testing.T) {
	commentMessage, err := json.Marshal(NotificationMessage{
		Template: "{actor} commented on your feedback",
		Variables: map[string]Variable{
			"actor": {Value: "Maya Chen", Type: "actor"},
		},
	})
	require.NoError(t, err)
	statusMessage, err := json.Marshal(NotificationMessage{
		Template: "{actor} marked your feedback as {status}",
		Variables: map[string]Variable{
			"actor":  {Value: "Sarah Jones", Type: "actor"},
			"status": {Value: "planned", Type: "value"},
		},
	})
	require.NoError(t, err)
	items := []NotificationEmailDigestItem{
		{
			NotificationID: uuid.New(),
			EntityType:     "feedback",
			FeedbackSlug:   "export-roadmap-to-pdf",
			Title:          "Export roadmap to PDF",
			Message:        commentMessage,
		},
		{
			NotificationID: uuid.New(),
			EntityType:     "feedback",
			FeedbackSlug:   "dark-mode",
			Title:          "Dark mode",
			Message:        statusMessage,
		},
	}

	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		WorkspaceName: "Product",
		Items:         items,
	}, "https://product.fortyone.app")

	require.NoError(t, err)
	require.Len(t, input.Request.Facts, 3)
	require.Contains(t, input.Request.Facts[1].Text, "commented on your feedback")
	require.Contains(t, input.Request.Facts[2].Text, "marked your feedback as")
	require.Equal(t, "https://product.fortyone.app/feedback/export-roadmap-to-pdf", input.Actions[input.FactActions[input.Request.Facts[1].ReferenceID]])
	require.Equal(t, "https://product.fortyone.app/feedback/dark-mode", input.Actions[input.FactActions[input.Request.Facts[2].ReferenceID]])
	require.True(t, feedbackOnlyDigest(items))
}

func TestBuildNotificationDigestCopyInputProtectsRenderedVariableFacts(t *testing.T) {
	message, err := json.Marshal(NotificationMessage{
		Template: "{actor} changed {field} to {value}",
		Variables: map[string]Variable{
			"actor": {Value: "Maya Chen", Type: "actor"},
			"field": {Value: "status", Type: "field"},
			"value": {Value: "In Progress", Type: "value"},
		},
	})
	require.NoError(t, err)

	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
		Items: []NotificationEmailDigestItem{{
			NotificationID: uuid.New(),
			EntityType:     "story",
			EntityID:       uuid.New(),
			Title:          "Make onboarding effortless",
			Message:        message,
		}},
	}, "https://product.fortyone.app")
	require.NoError(t, err)

	var notificationFact emailcopy.Fact
	for _, fact := range input.Request.Facts {
		if fact.ReferenceID == "notification_1" {
			notificationFact = fact
			break
		}
	}
	require.Equal(t, []string{"Maya Chen changed status to In Progress"}, notificationFact.ProtectedTokens)
}

func TestBuildNotificationDigestCopyInputProtectsSingleActorRole(t *testing.T) {
	message, err := json.Marshal(NotificationMessage{
		Template: "{actor} assigned you a task",
		Variables: map[string]Variable{
			"actor": {Value: "Maya Chen", Type: "actor"},
		},
	})
	require.NoError(t, err)

	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		WorkspaceName: "Product",
		Items: []NotificationEmailDigestItem{{
			NotificationID: uuid.New(),
			EntityType:     "story",
			EntityID:       uuid.New(),
			Title:          "Make onboarding effortless",
			Message:        message,
		}},
	}, "https://product.fortyone.app")
	require.NoError(t, err)

	var notificationFact emailcopy.Fact
	for _, fact := range input.Request.Facts {
		if fact.ReferenceID == "notification_1" {
			notificationFact = fact
			break
		}
	}
	require.Equal(t, []string{"Maya Chen assigned you a task"}, notificationFact.ProtectedTokens)
}

func TestNotificationSemanticProtectedTokensKeepsLongCommentAuthorRole(t *testing.T) {
	comment := strings.Repeat("The export flow needs a clearer progress state. ", 12)
	message := NotificationMessage{
		Template: "{actor} left a comment: " + comment,
		Variables: map[string]Variable{
			"actor": {Value: "Maya Chen", Type: "actor"},
		},
	}
	parsed := parseNotificationMessage(message)

	require.Equal(
		t,
		[]string{"Maya Chen left a comment"},
		notificationSemanticProtectedTokens(message, parsed.Text),
	)
}

func TestBuildNotificationDigestCopyInputExpandsWeeklyStrategyLinks(t *testing.T) {
	notificationID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	objectiveID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	keyResultID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	teamID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	health := "At Risk"
	currentValue := 42.0
	targetValue := 100.0
	rawMessage, err := json.Marshal(NotificationMessage{
		Template:  "Weekly strategy check-in",
		Variables: map[string]Variable{},
		Strategy: &strategyNotificationSnapshot{
			Version: 1,
			Kind:    "weekly_check_in",
			WeeklyCheckIn: &strategyWeeklyCheckInSnapshot{
				StaleAfterDays: 7,
				Counts: strategyWeeklyCheckInCounts{
					AtRiskObjectives: 1,
					StaleObjectives:  1,
					StaleKeyResults:  1,
					UniqueObjectives: 1,
				},
				Objectives: []strategyObjectiveSnapshot{{
					ID:        objectiveID,
					TeamID:    teamID,
					Name:      "Grow enterprise revenue",
					Health:    &health,
					UpdatedAt: time.Date(2026, time.August, 1, 8, 0, 0, 0, time.UTC),
					Reasons:   []string{"at_risk", "stale"},
				}},
				KeyResults: []strategyKeyResultSnapshot{{
					ID:              keyResultID,
					ObjectiveID:     objectiveID,
					TeamID:          teamID,
					Name:            "Reach $100k MRR",
					ObjectiveName:   "Grow enterprise revenue",
					MeasurementType: "number",
					CurrentValue:    &currentValue,
					TargetValue:     &targetValue,
					UpdatedAt:       time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC),
					Reasons:         []string{"stale", "incomplete"},
				}},
			},
		},
	})
	require.NoError(t, err)
	workspaceURL := "https://product.fortyone.app"
	input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
		WorkspaceName: "Product",
		Items: []NotificationEmailDigestItem{{
			NotificationID:   notificationID,
			NotificationType: "strategy_update",
			EntityType:       "strategy",
			Title:            "Your weekly strategy check-in",
			Message:          rawMessage,
		}},
	}, workspaceURL)

	require.NoError(t, err)
	require.True(t, input.HasStrategySnapshot)
	require.True(t, input.Request.IncludeSenderProse)
	require.False(t, input.Request.IncludeReplyPrompt)
	require.Len(t, input.Request.Facts, 4)
	require.Len(t, input.Request.Actions, 1)
	require.Equal(t, workspaceURL+"/strategy", input.Actions[digestActionStrategy])
	require.Equal(t, mailer.SenderProfileMaya, input.Fallback.Sender)
	require.Equal(t, "Your strategy check-in", input.Fallback.Subject)
	require.Contains(t, input.Fallback.Intro, "need your attention")

	objectiveFact := input.Request.Facts[2]
	keyResultFact := input.Request.Facts[3]
	require.Equal(t, "Grow enterprise revenue", objectiveFact.EntityTokens[0])
	require.Contains(t, objectiveFact.Text, "health is At Risk")
	require.Contains(t, keyResultFact.Text, "current value is 42 and target value is 100")
	require.True(t, objectiveFact.Required)
	require.True(t, keyResultFact.Required)

	objectiveURL := input.Actions[input.FactActions[objectiveFact.ReferenceID]]
	keyResultURL := input.Actions[input.FactActions[keyResultFact.ReferenceID]]
	require.Equal(
		t,
		workspaceURL+"/notifications/"+notificationID.String()+"?entityId="+objectiveID.String()+"&entityType=objective",
		objectiveURL,
	)
	require.Equal(
		t,
		workspaceURL+"/notifications/"+notificationID.String()+"?entityId="+keyResultID.String()+"&entityType=key_result&objectiveId="+objectiveID.String(),
		keyResultURL,
	)
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
			Intro:   "Here are the objectives, key results, and strategy updates that need your attention.",
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
