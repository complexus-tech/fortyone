package slack

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestBuildSlackObjectiveAndSprintUnfurlsKeepDetailsForExpandedView(t *testing.T) {
	t.Parallel()

	teamID := "11111111-1111-4111-8111-111111111111"
	objectiveID := "22222222-2222-4222-8222-222222222222"
	sprintID := "33333333-3333-4333-8333-333333333333"
	startDate := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)

	objectiveRequest, err := BuildSlackObjectiveUnfurlRequest("C123", "1754700000.123", SlackObjectiveWorkObjectInput{
		AccessGranted: true,
		ObjectiveURL:  "https://acme.fortyone.app/teams/" + teamID + "/objectives/" + objectiveID,
		Title:         "Improve onboarding",
		Description:   "Detailed objective context",
		Health:        "On Track",
		Progress:      "60% (6/10 stories)",
		LeadName:      "Maya Chen",
		StartDate:     &startDate,
		EndDate:       &endDate,
	})
	require.NoError(t, err)
	objectiveEntity := objectiveRequest.Metadata.Entities[0]
	require.Equal(t, slackObjectiveExternalRefType, objectiveEntity.ExternalRef.Type)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID+"/objectives/"+objectiveID, objectiveEntity.URL)
	require.NotContains(t, objectiveEntity.EntityPayload.Fields, "description")
	require.Equal(t, SlackWorkObjectField{Label: "Health", Value: "On Track"}, objectiveEntity.EntityPayload.Fields["status"])
	require.Equal(t, SlackWorkObjectField{
		Label: "Lead",
		Type:  slackUserFieldType,
		User:  &SlackWorkObjectUser{Text: "Maya Chen"},
	}, objectiveEntity.EntityPayload.Fields["assignee"])
	require.Equal(t, SlackWorkObjectField{
		Label: "End date",
		Value: "2026-08-31",
		Type:  slackDateFieldType,
	}, objectiveEntity.EntityPayload.Fields["due_date"])
	require.Equal(t, []SlackWorkObjectCustomField{
		{Key: "progress", Label: "Progress", Value: "60% (6/10 stories)", Type: "string"},
		{Key: "start_date", Label: "Start date", Value: "2026-08-01", Type: slackDateFieldType},
	}, objectiveEntity.EntityPayload.CustomFields)
	require.Equal(t, []string{"status", "progress", "assignee", "start_date", "due_date"}, objectiveEntity.EntityPayload.DisplayOrder)
	require.Equal(t, slackOpenObjectiveActionID, objectiveEntity.EntityPayload.Actions.PrimaryActions[0].ActionID)

	objectiveDetails, err := BuildSlackObjectiveEntityDetailsRequest("trigger-objective", SlackObjectiveWorkObjectInput{
		AccessGranted: true,
		ObjectiveURL:  objectiveEntity.URL,
		Title:         "Improve onboarding",
		Description:   "Detailed objective context",
	})
	require.NoError(t, err)
	require.Equal(t, "Detailed objective context", objectiveDetails.Metadata.EntityPayload.Fields["description"].Value)
	require.Equal(t, []string{"description"}, objectiveDetails.Metadata.EntityPayload.DisplayOrder)

	sprintRequest, err := BuildSlackSprintUnfurlRequest("C123", "1754700000.123", SlackSprintWorkObjectInput{
		AccessGranted: true,
		SprintURL:     "https://acme.fortyone.app/teams/" + teamID + "/sprints/" + sprintID + "/stories",
		Title:         "Sprint 12",
		Goal:          "Ship onboarding improvements",
		Status:        "Active",
		Progress:      "40% (4/10 stories)",
		StartDate:     &startDate,
		EndDate:       &endDate,
	})
	require.NoError(t, err)
	sprintEntity := sprintRequest.Metadata.Entities[0]
	require.Equal(t, slackSprintExternalRefType, sprintEntity.ExternalRef.Type)
	require.NotContains(t, sprintEntity.EntityPayload.Fields, "goal")
	require.Equal(t, "Active", sprintEntity.EntityPayload.Fields["status"].Value)
	require.Equal(t, []SlackWorkObjectCustomField{
		{Key: "progress", Label: "Progress", Value: "40% (4/10 stories)", Type: "string"},
		{Key: "start_date", Label: "Start date", Value: "2026-08-01", Type: slackDateFieldType},
		{Key: "end_date", Label: "End date", Value: "2026-08-31", Type: slackDateFieldType},
	}, sprintEntity.EntityPayload.CustomFields)
	require.Equal(t, []string{"status", "progress", "start_date", "end_date"}, sprintEntity.EntityPayload.DisplayOrder)
	require.Equal(t, slackOpenSprintActionID, sprintEntity.EntityPayload.Actions.PrimaryActions[0].ActionID)

	sprintDetails, err := BuildSlackSprintEntityDetailsRequest("trigger-sprint", SlackSprintWorkObjectInput{
		AccessGranted: true,
		SprintURL:     sprintEntity.URL,
		Title:         "Sprint 12",
		Goal:          "Ship onboarding improvements",
	})
	require.NoError(t, err)
	require.Equal(t, "Ship onboarding improvements", sprintDetails.Metadata.EntityPayload.Fields["goal"].Value)
}

func TestBuildSlackStoryUnfurlRequestRequiresAccessAndBuildsTaskMetadata(t *testing.T) {
	t.Parallel()

	dueDate := time.Date(2026, time.August, 19, 16, 30, 0, 0, time.FixedZone("CAT", 2*60*60))
	createdAt := time.Date(2026, time.August, 9, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)
	input := SlackStoryWorkObjectInput{
		AccessGranted:       true,
		StoryURL:            "https://acme.fortyone.app/work/web-123",
		Title:               "Fix workspace login",
		Description:         "<p>Users cannot sign in after accepting an invite.</p>",
		Status:              "In progress",
		StatusColor:         "blue",
		Priority:            "High",
		AssigneeSlackUserID: "U123ABC",
		CreatorName:         "Joseph Mukorivo",
		DueDate:             &dueDate,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}

	request, err := BuildSlackStoryUnfurlRequest("C123", "1754700000.123", input)
	require.NoError(t, err)
	require.NotNil(t, request.Metadata)
	require.Len(t, request.Metadata.Entities, 1)
	entity := request.Metadata.Entities[0]
	require.Equal(t, slackTaskEntityType, entity.EntityType)
	require.Equal(t, "https://acme.fortyone.app/work/web-123", entity.AppUnfurlURL)
	require.Equal(t, "https://acme.fortyone.app/work/WEB-123", entity.URL)
	require.Equal(t, "acme:WEB-123", entity.ExternalRef.ID)
	require.Equal(t, "Fix workspace login", entity.EntityPayload.Attributes.Title.Text)
	require.Equal(t, "WEB-123", entity.EntityPayload.Attributes.DisplayID)
	require.Equal(t, updatedAt.Unix(), entity.EntityPayload.Attributes.MetadataLastModified)
	require.Equal(t, "In progress", entity.EntityPayload.Fields["status"].Value)
	require.Equal(t, "blue", entity.EntityPayload.Fields["status"].TagColor)
	require.NotContains(t, entity.EntityPayload.Fields, "description")
	require.NotContains(t, entity.EntityPayload.Fields, "created_by")
	require.NotContains(t, entity.EntityPayload.Fields, "date_created")
	require.NotContains(t, entity.EntityPayload.Fields, "date_updated")
	require.Equal(t, "U123ABC", entity.EntityPayload.Fields["assignee"].User.UserID)
	require.Equal(t, "2026-08-19", entity.EntityPayload.Fields["due_date"].Value)
	require.Equal(t, slackDateFieldType, entity.EntityPayload.Fields["due_date"].Type)
	require.Equal(t, slackEditStoryStatusActionID, entity.EntityPayload.Actions.PrimaryActions[0].ActionID)
	require.Equal(t, slackEditStoryPriorityActionID, entity.EntityPayload.Actions.PrimaryActions[1].ActionID)
	require.Equal(t, slackOpenStoryActionID, entity.EntityPayload.Actions.OverflowActions[0].ActionID)

	input.AccessGranted = false
	input.Title = "A private story title"
	_, err = BuildSlackStoryUnfurlRequest("C123", "1754700000.123", input)
	require.ErrorIs(t, err, ErrSlackStoryPreviewAccessDenied)
}

func TestBuildSlackStoryEntityDetailsKeepsCompactPreviewFieldsForExpandedView(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.August, 9, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)
	request, err := BuildSlackStoryEntityDetailsRequest("trigger-123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
		Description:   "Users cannot sign in after accepting an invite.",
		CreatorName:   "Joseph Mukorivo",
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	})
	require.NoError(t, err)
	require.NotNil(t, request.Metadata)
	require.Equal(t, "Users cannot sign in after accepting an invite.", request.Metadata.EntityPayload.Fields["description"].Value)
	require.Equal(t, "Joseph Mukorivo", request.Metadata.EntityPayload.Fields["created_by"].User.Text)
	require.Equal(t, createdAt.Unix(), request.Metadata.EntityPayload.Fields["date_created"].Value)
	require.Equal(t, updatedAt.Unix(), request.Metadata.EntityPayload.Fields["date_updated"].Value)
}

func TestBuildSlackRequestUnfurlRequestRequiresAccessAndBuildsReadOnlyTaskMetadata(t *testing.T) {
	t.Parallel()

	dueDate := time.Date(2026, time.August, 19, 16, 30, 0, 0, time.FixedZone("CAT", 2*60*60))
	createdAt := time.Date(2026, time.August, 9, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)
	input := SlackRequestWorkObjectInput{
		AccessGranted:       true,
		RequestURL:          "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222?from=slack",
		Title:               "Fix workspace login",
		Description:         "<p>Users cannot sign in after accepting an invite.</p>",
		Status:              "Pending",
		Priority:            "High",
		AssigneeSlackUserID: "U123ABC",
		CreatorName:         "Joseph Mukorivo",
		DueDate:             &dueDate,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}

	request, err := BuildSlackRequestUnfurlRequest("C123", "1754700000.123", input)
	require.NoError(t, err)
	require.NotNil(t, request.Metadata)
	require.Len(t, request.Metadata.Entities, 1)
	entity := request.Metadata.Entities[0]
	require.Equal(t, slackTaskEntityType, entity.EntityType)
	require.Equal(t, slackRequestExternalRefType, entity.ExternalRef.Type)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222?from=slack",
		entity.AppUnfurlURL,
	)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		entity.URL,
	)
	require.Equal(t,
		"acme:11111111-1111-4111-8111-111111111111:22222222-2222-4222-8222-222222222222",
		entity.ExternalRef.ID,
	)
	require.Equal(t, "Fix workspace login", entity.EntityPayload.Attributes.Title.Text)
	require.Empty(t, entity.EntityPayload.Attributes.DisplayID)
	require.Nil(t, entity.EntityPayload.Attributes.Title.Edit)
	require.Equal(t, updatedAt.Unix(), entity.EntityPayload.Attributes.MetadataLastModified)
	require.Equal(t, "Pending", entity.EntityPayload.Fields["status"].Value)
	require.Equal(t, "Users cannot sign in after accepting an invite.", entity.EntityPayload.Fields["description"].Value)
	require.Equal(t, "U123ABC", entity.EntityPayload.Fields["assignee"].User.UserID)
	require.Equal(t, "2026-08-19", entity.EntityPayload.Fields["due_date"].Value)
	require.Equal(t, slackDateFieldType, entity.EntityPayload.Fields["due_date"].Type)
	for _, field := range entity.EntityPayload.Fields {
		require.Nil(t, field.Edit)
	}
	require.Equal(t, slackOpenRequestActionID, entity.EntityPayload.Actions.PrimaryActions[0].ActionID)
	require.Equal(t, "22222222-2222-4222-8222-222222222222", entity.EntityPayload.Actions.PrimaryActions[0].Value)

	input.AccessGranted = false
	input.Title = "A private request title"
	_, err = BuildSlackRequestUnfurlRequest("C123", "1754700000.123", input)
	require.ErrorIs(t, err, ErrSlackRequestPreviewAccessDenied)
}

func TestSlackWorkObjectDescriptionConvertsRichTextWithoutDamagingPlainText(t *testing.T) {
	t.Parallel()

	require.Equal(t,
		"First paragraph\n• One & only\n• Two\nFinal line",
		slackWorkObjectDescription(`<p>First <strong>paragraph</strong></p><ul><li>One &amp; only</li><li>Two</li></ul><div>Final line</div><script>private()</script>`),
	)
	require.Equal(t, "Use value <T> without conversion", slackWorkObjectDescription("Use value <T> without conversion"))
}

func TestBuildSlackStoryAuthenticationUnfurlDoesNotLeakStoryIdentity(t *testing.T) {
	t.Parallel()

	request, err := BuildSlackStoryAuthenticationUnfurlRequest(
		"C123",
		"1754700000.123",
		"https://acme.fortyone.app/settings/integrations/slack?slack_link_token=opaque",
	)
	require.NoError(t, err)
	require.Nil(t, request.Metadata)
	require.True(t, request.UserAuthRequired)
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "WEB-123")
	require.NotContains(t, string(encoded), "story")
}

func TestBuildSlackStoryCreationReceiptUsesExactTopLineAndDurableWorkObject(t *testing.T) {
	t.Parallel()

	receipt, err := BuildSlackStoryCreationReceipt("Joseph", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/web-123",
		Title:         "Fix workspace login",
		Status:        "Backlog",
		Description:   "A long story description that belongs in the expanded view.",
		CreatorName:   "Joseph Mukorivo",
		CreatedAt:     time.Date(2026, time.August, 9, 7, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, time.August, 9, 9, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.Equal(t, "Joseph created <https://acme.fortyone.app/work/WEB-123|WEB-123>", receipt.Text)
	require.NotContains(t, receipt.Text, "✅")
	require.NotContains(t, receipt.Text, "FortyOne")
	require.NotNil(t, receipt.ProviderPayload.Metadata)
	entity := receipt.ProviderPayload.Metadata.Entities[0]
	require.Len(t, entity.EntityPayload.Actions.PrimaryActions, 1)
	require.NotContains(t, entity.EntityPayload.Fields, "description")
	require.NotContains(t, entity.EntityPayload.Fields, "created_by")
	require.NotContains(t, entity.EntityPayload.Fields, "date_created")
	require.NotContains(t, entity.EntityPayload.Fields, "date_updated")
	require.Empty(t, receipt.ProviderPayload.Metadata.Entities[0].AppUnfurlURL)

	encoded, err := EncodeSlackProviderPayload(receipt.ProviderPayload)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"unfurl_links":false`)
	require.Contains(t, string(encoded), `"unfurl_media":false`)
	restored, err := DecodeSlackProviderPayload(encoded)
	require.NoError(t, err)
	require.Equal(t, receipt.ProviderPayload, restored)
}

func TestBuildSlackRequestCreationReceiptUsesLinkedOpeningCopyAndDurableWorkObject(t *testing.T) {
	t.Parallel()

	receipt, err := BuildSlackRequestCreationReceipt("Joseph", SlackRequestWorkObjectInput{
		AccessGranted: true,
		RequestURL:    "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		Title:         "Fix workspace login",
		Status:        "Pending",
	})
	require.NoError(t, err)
	require.Equal(t,
		"Joseph <https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222|opened a request>",
		receipt.Text,
	)
	require.NotContains(t, receipt.Text, "📥")
	require.NotContains(t, receipt.Text, "in FortyOne")
	require.NotNil(t, receipt.ProviderPayload.Metadata)
	entity := receipt.ProviderPayload.Metadata.Entities[0]
	require.Equal(t, slackRequestExternalRefType, entity.ExternalRef.Type)
	require.Empty(t, entity.AppUnfurlURL)

	encoded, err := EncodeSlackProviderPayload(receipt.ProviderPayload)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"unfurl_links":false`)
	require.Contains(t, string(encoded), `"unfurl_media":false`)
	restored, err := DecodeSlackProviderPayload(encoded)
	require.NoError(t, err)
	require.Equal(t, receipt.ProviderPayload, restored)
}

func TestBuildSlackRequestCreationReceiptEscapesActorAndFallsBack(t *testing.T) {
	t.Parallel()

	input := SlackRequestWorkObjectInput{
		AccessGranted: true,
		RequestURL:    "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		Title:         "Fix workspace login",
	}
	receipt, err := BuildSlackRequestCreationReceipt(" Joseph  <@U123> & Team ", input)
	require.NoError(t, err)
	require.Equal(t,
		"Joseph &lt;@U123&gt; &amp; Team <https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222|opened a request>",
		receipt.Text,
	)

	receipt, err = BuildSlackRequestCreationReceipt("   ", input)
	require.NoError(t, err)
	require.Equal(t,
		"A team member <https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222|opened a request>",
		receipt.Text,
	)
}

func TestBuildSlackRequestEntityDetailsRequestOmitsAppUnfurlURL(t *testing.T) {
	t.Parallel()

	request, err := BuildSlackRequestEntityDetailsRequest(" trigger-123 ", SlackRequestWorkObjectInput{
		AccessGranted: true,
		RequestURL:    "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222?from=slack",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	require.Equal(t, "trigger-123", request.TriggerID)
	require.NotNil(t, request.Metadata)
	require.Equal(t, slackRequestExternalRefType, request.Metadata.ExternalRef.Type)
	require.Empty(t, request.Metadata.AppUnfurlURL)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		request.Metadata.URL,
	)
}

func TestBuildSlackMutationConfirmationProviderPayloadUsesOpaqueButtonValues(t *testing.T) {
	t.Parallel()

	payload, err := BuildSlackMutationConfirmationProviderPayload("Create *WEB-123* in Backlog?", "opaque-confirmation-token", "U123", false)
	require.NoError(t, err)
	require.Len(t, payload.Blocks, 2)
	require.Equal(t, "section", payload.Blocks[0].Type)
	require.Equal(t, "actions", payload.Blocks[1].Type)
	require.Equal(t, slackConfirmMutationActionID, payload.Blocks[1].Elements[0].ActionID)
	require.Equal(t, slackCancelMutationActionID, payload.Blocks[1].Elements[1].ActionID)
	for _, element := range payload.Blocks[1].Elements {
		value, decodeErr := decodeSlackMutationActionValue(element.Value)
		require.NoError(t, decodeErr)
		require.Equal(t, "U123", value.SlackUserID)
		require.Equal(t, "opaque-confirmation-token", value.Token)
	}
	encoded, err := EncodeSlackProviderPayload(payload)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "👍")
	_, err = DecodeSlackProviderPayload(append(encoded, []byte(` {}`)...))
	require.Error(t, err)
}

func TestBuildSlackMutationConfirmationProviderPayloadLabelsBatchActionCreateAll(t *testing.T) {
	t.Parallel()

	payload, err := BuildSlackMutationConfirmationProviderPayload(
		"Create 2 stories?",
		"opaque-batch-token",
		"U123",
		true,
	)
	require.NoError(t, err)
	require.Len(t, payload.Blocks, 2)
	require.Equal(t, "Create all", payload.Blocks[1].Elements[0].Text.Text)
	require.Equal(t, "Create all proposed stories", payload.Blocks[1].Elements[0].AccessibilityLabel)
	require.Equal(t, "Cancel", payload.Blocks[1].Elements[1].Text.Text)
}

func TestBuildSlackMutationConfirmationProviderPayloadPreservesLongBatchPrompt(t *testing.T) {
	t.Parallel()

	prompt := strings.Join([]string{
		"Create 10 stories?",
		strings.Repeat("a", 2_000),
		strings.Repeat("b", 2_000),
		strings.Repeat("c", 2_000),
	}, "\n")
	payload, err := BuildSlackMutationConfirmationProviderPayload(
		prompt,
		"opaque-batch-token",
		"U123",
		true,
	)
	require.NoError(t, err)
	require.Len(t, payload.Blocks, 4)
	require.Equal(t, "actions", payload.Blocks[len(payload.Blocks)-1].Type)

	sections := make([]string, 0, len(payload.Blocks)-1)
	for _, block := range payload.Blocks[:len(payload.Blocks)-1] {
		require.Equal(t, "section", block.Type)
		require.NotNil(t, block.Text)
		require.LessOrEqual(t, utf8.RuneCountInString(block.Text.Text), slackWorkObjectTextFieldLimit)
		sections = append(sections, block.Text.Text)
	}
	require.Equal(t, prompt, strings.Join(sections, "\n"))
}

func TestBuildSlackMutationRetryProviderPayloadHasOnlyRetryAction(t *testing.T) {
	t.Parallel()

	payload, err := BuildSlackMutationRetryProviderPayload(
		"One story was created. Retry the remaining stories?",
		"opaque-batch-token",
		"U123",
	)
	require.NoError(t, err)
	require.Len(t, payload.Blocks, 2)
	require.Len(t, payload.Blocks[1].Elements, 1)
	retry := payload.Blocks[1].Elements[0]
	require.Equal(t, slackConfirmMutationActionID, retry.ActionID)
	require.Equal(t, "Retry remaining", retry.Text.Text)
	require.Equal(t, "Retry creating the remaining proposed stories", retry.AccessibilityLabel)
	value, err := decodeSlackMutationActionValue(retry.Value)
	require.NoError(t, err)
	require.Equal(t, "U123", value.SlackUserID)
	require.Equal(t, "opaque-batch-token", value.Token)
	_, err = EncodeSlackProviderPayload(payload)
	require.NoError(t, err)
}
