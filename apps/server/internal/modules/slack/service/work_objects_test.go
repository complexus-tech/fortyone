package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParseFortyOneStoryURL(t *testing.T) {
	t.Parallel()

	link, err := ParseFortyOneStoryURL("https://acme.fortyone.app/work/web-123?from=slack")
	require.NoError(t, err)
	require.Equal(t, "acme", link.WorkspaceSlug)
	require.Equal(t, "WEB-123", link.StoryReference)
	require.Equal(t, "https://acme.fortyone.app/work/WEB-123", link.CanonicalURL)
	require.Equal(t, "https://acme.fortyone.app/work/web-123?from=slack", link.PostedURL)
}

func TestParseFortyOneStoryURLRejectsUntrustedOrNonStoryURLs(t *testing.T) {
	t.Parallel()

	invalidURLs := []string{
		"http://acme.fortyone.app/work/WEB-123",
		"https://fortyone.app/work/WEB-123",
		"https://acme.fortyone.app.evil.example/work/WEB-123",
		"https://one.two.fortyone.app/work/WEB-123",
		"https://acme.fortyone.app:8443/work/WEB-123",
		"https://user@acme.fortyone.app/work/WEB-123",
		"https://acme.fortyone.app/feedback/WEB-123",
		"https://acme.fortyone.app/work/WEB-0",
		"https://acme.fortyone.app/work/WEB-123%2Fprivate",
	}
	for _, rawURL := range invalidURLs {
		rawURL := rawURL
		t.Run(rawURL, func(t *testing.T) {
			t.Parallel()
			_, err := ParseFortyOneStoryURL(rawURL)
			require.ErrorIs(t, err, ErrInvalidFortyOneStoryURL)
		})
	}
}

func TestParseFortyOneRequestURL(t *testing.T) {
	t.Parallel()

	teamID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	requestID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	link, err := ParseFortyOneRequestURL(
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222?from=slack",
	)
	require.NoError(t, err)
	require.Equal(t, "acme", link.WorkspaceSlug)
	require.Equal(t, teamID, link.TeamID)
	require.Equal(t, requestID, link.RequestID)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		link.CanonicalURL,
	)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222?from=slack",
		link.PostedURL,
	)

	trailingSlashLink, err := ParseFortyOneRequestURL(
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222/",
	)
	require.NoError(t, err)
	require.Equal(t, link.CanonicalURL, trailingSlashLink.CanonicalURL)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222/",
		trailingSlashLink.PostedURL,
	)
}

func TestParseFortyOneRequestURLRejectsUntrustedOrNonCanonicalURLs(t *testing.T) {
	t.Parallel()

	const canonicalPath = "/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222"
	invalidURLs := []string{
		"http://acme.fortyone.app" + canonicalPath,
		"https://fortyone.app" + canonicalPath,
		"https://acme.fortyone.app.evil.example" + canonicalPath,
		"https://one.two.fortyone.app" + canonicalPath,
		"https://acme.fortyone.app:8443" + canonicalPath,
		"https://user@acme.fortyone.app" + canonicalPath,
		"https://acme.fortyone.app/feedback/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/not-a-uuid/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/00000000-0000-0000-0000-000000000000/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/not-a-uuid",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/00000000-0000-0000-0000-000000000000",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222/activity",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222//",
		"https://acme.fortyone.app//teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/Teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/11111111111141118111111111111111/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/{11111111-1111-4111-8111-111111111111}/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/urn:uuid:11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222%2Fprivate",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222%5Cprivate",
		"https://acme.fortyone.app/teams/11111111%2D1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
	}
	for _, rawURL := range invalidURLs {
		rawURL := rawURL
		t.Run(rawURL, func(t *testing.T) {
			t.Parallel()
			_, err := ParseFortyOneRequestURL(rawURL)
			require.ErrorIs(t, err, ErrInvalidFortyOneRequestURL)
		})
	}
}

func TestParseFortyOneObjectiveAndSprintURLs(t *testing.T) {
	t.Parallel()

	teamID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	objectiveID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	sprintID := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	objective, err := ParseFortyOneObjectiveURL("https://acme.fortyone.app/teams/" + teamID.String() + "/objectives/" + objectiveID.String() + "?from=slack")
	require.NoError(t, err)
	require.Equal(t, teamID, objective.TeamID)
	require.Equal(t, objectiveID, objective.ObjectiveID)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID.String()+"/objectives/"+objectiveID.String(), objective.CanonicalURL)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID.String()+"/objectives/"+objectiveID.String()+"?from=slack", objective.PostedURL)

	sprint, err := ParseFortyOneSprintURL("https://acme.fortyone.app/teams/" + teamID.String() + "/sprints/" + sprintID.String() + "/stories/")
	require.NoError(t, err)
	require.Equal(t, teamID, sprint.TeamID)
	require.Equal(t, sprintID, sprint.SprintID)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID.String()+"/sprints/"+sprintID.String()+"/stories", sprint.CanonicalURL)
}

func TestParseFortyOneObjectiveAndSprintURLsRejectInvalidRoutes(t *testing.T) {
	t.Parallel()

	invalidURLs := []struct {
		name string
		url  string
		err  error
	}{
		{
			name: "objective wrong route",
			url:  "https://acme.fortyone.app/objectives/22222222-2222-4222-8222-222222222222",
			err:  ErrInvalidFortyOneObjectiveURL,
		},
		{
			name: "objective encoded path",
			url:  "https://acme.fortyone.app/teams/11111111%2D1111-4111-8111-111111111111/objectives/22222222-2222-4222-8222-222222222222",
			err:  ErrInvalidFortyOneObjectiveURL,
		},
		{
			name: "sprint missing stories route",
			url:  "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/sprints/33333333-3333-4333-8333-333333333333",
			err:  ErrInvalidFortyOneSprintURL,
		},
		{
			name: "sprint wrong trailing route",
			url:  "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/sprints/33333333-3333-4333-8333-333333333333/analytics",
			err:  ErrInvalidFortyOneSprintURL,
		},
	}
	for _, testCase := range invalidURLs {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if testCase.err == ErrInvalidFortyOneObjectiveURL {
				_, err := ParseFortyOneObjectiveURL(testCase.url)
				require.ErrorIs(t, err, testCase.err)
				return
			}
			_, err := ParseFortyOneSprintURL(testCase.url)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

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
		StartDate:     &startDate,
		EndDate:       &endDate,
	})
	require.NoError(t, err)
	objectiveEntity := objectiveRequest.Metadata.Entities[0]
	require.Equal(t, slackObjectiveExternalRefType, objectiveEntity.ExternalRef.Type)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID+"/objectives/"+objectiveID, objectiveEntity.URL)
	require.NotContains(t, objectiveEntity.EntityPayload.Fields, "description")
	require.Equal(t, "60% (6/10 stories)", objectiveEntity.EntityPayload.Fields["progress"].Value)
	require.Equal(t, slackOpenObjectiveActionID, objectiveEntity.EntityPayload.Actions.PrimaryActions[0].ActionID)

	objectiveDetails, err := BuildSlackObjectiveEntityDetailsRequest("trigger-objective", SlackObjectiveWorkObjectInput{
		AccessGranted: true,
		ObjectiveURL:  objectiveEntity.URL,
		Title:         "Improve onboarding",
		Description:   "Detailed objective context",
	})
	require.NoError(t, err)
	require.Equal(t, "Detailed objective context", objectiveDetails.Metadata.EntityPayload.Fields["description"].Value)

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

	payload, err := BuildSlackMutationConfirmationProviderPayload("Create *WEB-123* in Backlog?", "opaque-confirmation-token", "U123")
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

func TestSlackWorkObjectPublisherUnfurlsTypedMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, "/chat.unfurl", req.URL.Path)
		require.Equal(t, "Bearer xoxb-test", req.Header.Get("Authorization"))
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), `"entity_type":"slack#/entities/task"`)
		require.NotContains(t, string(body), "xoxb-test")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	request, err := BuildSlackStoryUnfurlRequest("C123", "1754700000.123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	require.NoError(t, publisher.Unfurl(context.Background(), "xoxb-test", request))
}

func TestSlackWorkObjectPublisherUsesComposerDestination(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var payload map[string]any
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
		require.Equal(t, "unfurl-123", payload["unfurl_id"])
		require.Equal(t, "composer", payload["source"])
		require.NotContains(t, payload, "channel")
		require.NotContains(t, payload, "ts")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	request, err := BuildSlackStoryUnfurlRequest("COMPOSER", "draft-ts", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	request.Channel = ""
	request.TS = ""
	request.UnfurlID = "unfurl-123"
	request.Source = "composer"
	require.NoError(t, publisher.Unfurl(context.Background(), "xoxb-test", request))
}

func TestPublishSlackStoryUnfurlLogsProviderErrorCode(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false,"error":"cannot_unfurl_url"}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	var logs bytes.Buffer
	processor := &EventProcessor{
		log:         logger.NewWithJSON(&logs, slog.LevelInfo, "test"),
		workObjects: newSlackWorkObjectPublisher(client),
	}
	request, err := BuildSlackStoryUnfurlRequest("C123", "1754700000.123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://complexus.fortyone.app/work/WEB-544",
		Title:         "Private story title that must not be logged",
	})
	require.NoError(t, err)

	err = processor.publishSlackStoryUnfurl(context.Background(), normalizedSlackEvent{
		EventID:   "Ev-preview",
		TeamID:    "T123",
		UserID:    "U123",
		ChannelID: "C123",
		MessageTS: "1754700000.123",
	}, uuid.MustParse("11111111-1111-4111-8111-111111111111"), request, 1, false, "xoxb-test")

	require.Error(t, err)
	require.Contains(t, logs.String(), `"msg":"Slack story preview publish failed"`)
	require.Contains(t, logs.String(), `"slack_error_code":"cannot_unfurl_url"`)
	require.Contains(t, logs.String(), `"event_id":"Ev-preview"`)
	require.NotContains(t, logs.String(), "Private story title that must not be logged")
}

func TestApplySlackUnfurlEventDestinationUsesComposerIdentity(t *testing.T) {
	t.Parallel()

	request := SlackChatUnfurlRequest{Channel: "COMPOSER", TS: "draft-ts"}
	applySlackUnfurlEventDestination(&request, normalizedSlackEvent{
		Source:   "composer",
		UnfurlID: "unfurl-123",
	})

	require.Empty(t, request.Channel)
	require.Empty(t, request.TS)
	require.Equal(t, "unfurl-123", request.UnfurlID)
	require.Equal(t, "composer", request.Source)
	require.NoError(t, validateSlackUnfurlRequestDestination(request))
}

func TestValidateSlackUnfurlRequestDestinationRejectsMixedPairs(t *testing.T) {
	t.Parallel()

	err := validateSlackUnfurlRequestDestination(SlackChatUnfurlRequest{
		Channel:  "C123",
		TS:       "1754700000.123",
		UnfurlID: "unfurl-123",
		Source:   "composer",
	})
	require.ErrorContains(t, err, "exactly one destination pair")
}

func TestSlackWorkObjectPublisherPresentsEntityDetailsWithoutAppUnfurlURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, "/entity.presentDetails", req.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
		require.Equal(t, "trigger-123", payload["trigger_id"])
		metadata, ok := payload["metadata"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, slackTaskEntityType, metadata["entity_type"])
		require.NotContains(t, metadata, "app_unfurl_url")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	request, err := BuildSlackStoryEntityDetailsRequest("trigger-123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	require.NoError(t, publisher.PresentDetails(context.Background(), "xoxb-test", request))
}

func TestSlackWorkObjectPublisherPostsRichCreationReceipt(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, "/chat.postMessage", req.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
		require.Equal(t, "Joseph created <https://acme.fortyone.app/work/WEB-123|WEB-123>", payload["text"])
		require.Equal(t, false, payload["unfurl_links"])
		require.NotNil(t, payload["metadata"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"ts":"1754700000.456"}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	receipt, err := BuildSlackStoryCreationReceipt("Joseph", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	externalMessageID, err := publisher.PostCreationReceipt(context.Background(), "xoxb-test", "C123", "", "client-id", receipt)
	require.NoError(t, err)
	require.Equal(t, "1754700000.456", externalMessageID)
}

func TestSlackWorkObjectPublisherRejectsMixedAuthAndMetadata(t *testing.T) {
	t.Parallel()

	publisher := newSlackWorkObjectPublisher(newSlackWebClient(http.DefaultClient))
	request, err := BuildSlackStoryUnfurlRequest("C123", "1754700000.123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Private title",
	})
	require.NoError(t, err)
	request.UserAuthRequired = true
	request.UserAuthURL = "https://acme.fortyone.app/settings/integrations/slack"
	err = publisher.Unfurl(context.Background(), "xoxb-test", request)
	require.ErrorContains(t, err, "cannot mix")
}

func TestSlackBotOAuthScopeValueIncludesChannelInventoryAndLinkScopes(t *testing.T) {
	t.Parallel()

	require.Equal(t,
		"app_mentions:read,channels:history,channels:read,chat:write,chat:write.public,commands,groups:history,groups:read,im:history,links:read,links:write",
		slackBotOAuthScopeValue(),
	)
}
