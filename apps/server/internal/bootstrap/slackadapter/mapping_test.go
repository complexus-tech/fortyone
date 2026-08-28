package slackadapter

import (
	"errors"
	"net/http"
	"testing"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIntegrationRequestMappingDoesNotRetainMutableCallerState(t *testing.T) {
	t.Parallel()

	labelID := uuid.New()
	request := integrationrequests.CoreIntegrationRequest{
		ID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New(),
		Provider: integrationrequests.ProviderSlack, Title: "Customer request",
		LabelIDs: []uuid.UUID{labelID}, Metadata: map[string]any{"slack_channel_id": "C1"},
	}
	mapped := mapIntegrationRequest(request)

	request.LabelIDs[0] = uuid.New()
	request.Metadata["slack_channel_id"] = "C2"
	require.Equal(t, labelID, mapped.LabelIDs[0])
	require.Equal(t, "C1", mapped.Metadata["slack_channel_id"])
}

func TestProviderThreadErrorsAreNormalizedAtTheAdapterBoundary(t *testing.T) {
	t.Parallel()

	err := mapRequestError(errors.Join(errors.New("query provider thread"), integrationrequests.ErrProviderThreadNotFound))
	require.ErrorIs(t, err, slack.ErrProviderThreadNotFound)
	require.ErrorIs(t, err, integrationrequests.ErrProviderThreadNotFound)
}

func TestStoryMappingAndErrorProjectionPreserveSlackContract(t *testing.T) {
	t.Parallel()

	storyID, workspaceID, teamID := uuid.New(), uuid.New(), uuid.New()
	description, descriptionHTML := "Details", "<p>Details</p>"
	createdAt := time.Date(2026, time.August, 28, 10, 0, 0, 0, time.UTC)
	mapped := mapStory(stories.CoreSingleStory{
		ID: storyID, SequenceID: 41, Workspace: workspaceID, Team: teamID,
		TeamCode: "API", Title: "Typed Slack boundary", Description: &description,
		DescriptionHTML: &descriptionHTML, Priority: "High", CreatedAt: createdAt, CreatedNow: true,
	})

	require.Equal(t, storyID, mapped.ID)
	require.Equal(t, workspaceID, mapped.Workspace)
	require.Equal(t, teamID, mapped.Team)
	require.Equal(t, "API", mapped.TeamCode)
	require.Equal(t, &descriptionHTML, mapped.DescriptionHTML)
	require.True(t, mapped.CreatedNow)
	require.ErrorIs(t, mapStoryError(stories.ErrStoryChanged), slack.ErrStoryChanged)
	require.ErrorIs(t, mapStoryError(stories.ErrNotFound), slack.ErrStoryNotFound)
}

func TestAssistantResponseAndProviderErrorsUseSlackOwnedValues(t *testing.T) {
	t.Parallel()

	response := mapAssistantResponse(messaging.Response{
		Text:  "Ready",
		Usage: messagingdomain.Usage{InputTokens: 12, OutputTokens: 4, TotalTokens: 16},
		Confirmation: &messaging.StoryMutationConfirmation{
			Operation: messaging.StoryMutationCreateBatch,
			Token:     "confirmation-token", Prompt: "Create two stories?",
		},
	})
	require.Equal(t, "Ready", response.Text)
	require.Equal(t, slack.AssistantUsage{InputTokens: 12, OutputTokens: 4, TotalTokens: 16}, response.Usage)
	require.Equal(t, slack.StoryMutationCreateBatch, response.Confirmation.Operation)

	providerErr := mapAssistantError(&messaging.APIError{
		StatusCode: http.StatusBadRequest, Code: "invalid_request_error",
		Message: "invalid model input", RequestID: "req_123",
	})
	var projected *slack.AssistantAPIError
	require.ErrorAs(t, providerErr, &projected)
	require.True(t, projected.Permanent)
	require.Equal(t, "invalid_request_error", projected.Code)
	require.Equal(t, "req_123", projected.RequestID)

	require.ErrorIs(t, mapAssistantError(messaging.ErrAssistantNotConfigured), slack.ErrAssistantNotConfigured)
}

func TestMessagingRecordsAndErrorsAreProjectedWithoutSharingBuffers(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"authorization":{"scope":"actor_membership"}}`)
	record := messagingrepository.OutboundDeliveryRecord{
		ID: uuid.New(), WorkspaceID: uuid.New(), ExternalWorkspaceID: "T1",
		ExternalChannelID: "C1", ProviderPayload: payload, Status: "delivering", AttemptCount: 2,
	}
	mapped := mapOutboundDelivery(record)
	payload[0] = '['
	require.Equal(t, byte('{'), mapped.ProviderPayload[0])
	require.Equal(t, 2, mapped.AttemptCount)

	require.ErrorIs(t, mapMessagingError(messagingrepository.ErrNotFound), slack.ErrMessagingRecordNotFound)
	require.ErrorIs(t, mapMessagingError(messagingrepository.ErrLeaseBusy), slack.ErrOutboundDeliveryBusy)
}

func TestMutationResultsKeepReceiptFieldsAndNormalizeFailures(t *testing.T) {
	t.Parallel()

	teamID, storyID := uuid.New(), uuid.New()
	mapped := mapMutationResult(messaging.StoryMutationResult{
		Status: "partial", Operation: messaging.StoryMutationCreateBatch, TeamID: teamID,
		Items: []messaging.StoryMutationItemResult{{
			Index: 0, Status: "applied", StoryID: storyID, TeamID: teamID,
			Reference: "API-41", Title: "Typed Slack boundary", Priority: "High",
		}},
	})
	require.Equal(t, slack.StoryMutationCreateBatch, mapped.Operation)
	require.Equal(t, "API-41", mapped.Items[0].Reference)
	require.Equal(t, "High", mapped.Items[0].Priority)
	require.ErrorIs(t, mapMutationError(messaging.ErrTeamNotAccessible), slack.ErrStoryMutationTeamRestricted)
}
