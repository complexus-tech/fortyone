package taskhandlers

import (
	"context"
	"encoding/json"
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

func TestNotificationGuidanceThreadContextIncludesStrategyTargets(t *testing.T) {
	objectiveID := uuid.New()
	keyResultID := uuid.New()
	teamID := uuid.New()
	message, err := json.Marshal(NotificationMessage{Strategy: &strategyNotificationSnapshot{
		Version: 1,
		Kind:    "weekly_check_in",
		WeeklyCheckIn: &strategyWeeklyCheckInSnapshot{
			Objectives: []strategyObjectiveSnapshot{{ID: objectiveID, TeamID: teamID, Name: "Reach the right customers"}},
			KeyResults: []strategyKeyResultSnapshot{{ID: keyResultID, ObjectiveID: objectiveID, TeamID: teamID, Name: "Increase qualified activation"}},
		},
	}})
	require.NoError(t, err)

	encoded, err := notificationGuidanceThreadContext(NotificationEmailDigestData{
		WorkspaceSlug: "product",
		Items:         []NotificationEmailDigestItem{{Message: message}},
	})
	require.NoError(t, err)
	require.Contains(t, string(encoded), objectiveID.String())
	require.Contains(t, string(encoded), keyResultID.String())
	require.Contains(t, string(encoded), `"source":"strategy_notification"`)
	require.Contains(t, string(encoded), `"parentId":"`+objectiveID.String()+`"`)
}

func TestNotificationDigestMessageIDIsStableAcrossItemOrder(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	base := NotificationEmailDigestData{WorkspaceID: uuid.New(), RecipientID: uuid.New()}
	base.Items = []NotificationEmailDigestItem{{NotificationID: firstID}, {NotificationID: secondID}}
	first := notificationDigestMessageID(base)

	base.Items = []NotificationEmailDigestItem{{NotificationID: secondID}, {NotificationID: firstID}}
	require.Equal(t, first, notificationDigestMessageID(base))
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
	require.True(t, input.Request.IncludeReplyPrompt)
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
