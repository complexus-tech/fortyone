package slack

import (
	"context"
	"errors"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
)

var (
	ErrProviderThreadNotFound      = errors.New("provider thread not found")
	ErrMessagingRecordNotFound     = errors.New("messaging record not found")
	ErrOutboundDeliveryBusy        = errors.New("outbound delivery lease is busy")
	ErrDailyWorkspaceTokenLimit    = errors.New("daily workspace assistant token limit reached")
	ErrAssistantNotConfigured      = errors.New("assistant is not configured")
	ErrAssistantPromptTooLarge     = errors.New("assistant message is too large")
	ErrStoryMutationApplied        = errors.New("story mutation confirmation was already applied")
	ErrStoryMutationCancelled      = errors.New("story mutation confirmation was cancelled")
	ErrStoryMutationExpired        = errors.New("story mutation confirmation expired")
	ErrStoryMutationInvalid        = errors.New("invalid story mutation confirmation")
	ErrStoryMutationNotAllowed     = errors.New("story mutations are not allowed")
	ErrStoryMutationStale          = errors.New("story changed after mutation confirmation was requested")
	ErrStoryMutationTeamRestricted = errors.New("team is not accessible to the user")
	ErrStoryNotFound               = errors.New("story not found")
	ErrInvalidStoryReference       = errors.New("invalid story reference")
	ErrStoryChanged                = errors.New("story changed before the update was applied")
)

type Assistant interface {
	Respond(ctx context.Context, request AssistantRequest) (AssistantResponse, error)
}

type StoryMutationConfirmer interface {
	ConfirmStoryMutation(ctx context.Context, scope StoryMutationScope, token string) (StoryMutationResult, error)
	CancelStoryMutation(ctx context.Context, scope StoryMutationScope, token string) (StoryMutationCancellationResult, error)
}

// Lower-case aliases keep the implementation concise while the exported
// contract remains available to composition adapters and focused test stubs.
type (
	integrationRequest            = IntegrationRequest
	providerThreadLookupInput     = ProviderThreadLookupInput
	upsertIntegrationRequestInput = UpsertIntegrationRequestInput
	conversationInput             = ConversationInput
	conversationRecord            = ConversationRecord
	dailyUsageRecordInput         = DailyUsageRecordInput
	dailyUsageSnapshot            = DailyUsageSnapshot
	messageRecord                 = MessageRecord
	nonceConsumeInput             = NonceConsumeInput
	nonceInput                    = NonceInput
	nonceRecord                   = NonceRecord
	outboundDeliveryInput         = OutboundDeliveryInput
	outboundDeliveryRecord        = OutboundDeliveryRecord
	assistant                     = Assistant
	assistantAPIError             = AssistantAPIError
	assistantConversationTurn     = AssistantConversationTurn
	assistantRequest              = AssistantRequest
	assistantRuntimeContext       = AssistantRuntimeContext
	assistantRuntimeSurface       = AssistantSurfaceContext
	storyMutationConfirmer        = StoryMutationConfirmer
	storyMutationItemResult       = StoryMutationItemResult
	storyMutationOperation        = StoryMutationOperation
	storyMutationScope            = StoryMutationScope
	objective                     = Objective
	assistantChannelTeamScope     = slackdomain.AssistantChannelTeamScope
	slackAgentSettingsRecord      = slackdomain.AgentSettings
	slackChannelRecord            = slackdomain.Channel
	slackChannelPayload           = slackdomain.ChannelUpsert
	slackLabelRecord              = slackdomain.Label
	slackOAuthInstallPayload      = slackdomain.OAuthInstallation
	slackRequestLogInsert         = slackdomain.RequestLogInsert
	slackRequestLogRecord         = slackdomain.RequestLog
	slackStatusRecord             = slackdomain.Status
	slackTeamMemberRecord         = slackdomain.TeamMember
	slackTeamRecord               = slackdomain.Team
	slackUninstallInput           = slackdomain.EnqueueUninstall
	slackUninstallRecord          = slackdomain.Uninstall
	slackUserLinkUpsert           = slackdomain.UserLinkUpsert
	slackWorkspaceRecord          = slackdomain.Installation
	workspaceRecord               = slackdomain.Workspace
	sprint                        = Sprint
	newStory                      = NewStory
	singleStory                   = Story
)

const (
	providerSlack                    = ProviderSlack
	integrationRequestStatusPending  = IntegrationRequestStatusPending
	integrationRequestStatusAccepted = IntegrationRequestStatusAccepted
	integrationRequestStatusDeclined = IntegrationRequestStatusDeclined
	conversationAudienceActor        = ConversationAudienceActor
	conversationAudienceChannel      = ConversationAudienceChannel
	assistantRoleUser                = AssistantRoleUser
	assistantRoleAssistant           = AssistantRoleAssistant
	assistantSurfaceDirect           = AssistantSurfaceDirect
	assistantSurfaceThread           = AssistantSurfaceThread
	storyMutationCreate              = StoryMutationCreate
	storyMutationCreateBatch         = StoryMutationCreateBatch
)

var (
	errProviderThreadNotFound      = ErrProviderThreadNotFound
	errDailyWorkspaceTokenLimit    = ErrDailyWorkspaceTokenLimit
	errMessagingRecordNotFound     = ErrMessagingRecordNotFound
	errAssistantNotConfigured      = ErrAssistantNotConfigured
	errAssistantPromptTooLarge     = ErrAssistantPromptTooLarge
	errStoryMutationApplied        = ErrStoryMutationApplied
	errStoryMutationCancelled      = ErrStoryMutationCancelled
	errStoryMutationExpired        = ErrStoryMutationExpired
	errStoryMutationInvalid        = ErrStoryMutationInvalid
	errStoryMutationNotAllowed     = ErrStoryMutationNotAllowed
	errStoryMutationStale          = ErrStoryMutationStale
	errStoryMutationTeamRestricted = ErrStoryMutationTeamRestricted
	errWorkspaceAlreadyConnected   = slackdomain.ErrWorkspaceAlreadyConnected
	errStoryNotFound               = ErrStoryNotFound
	errInvalidStoryReference       = ErrInvalidStoryReference
)

func isSlackRepositoryNotFound(err error) bool {
	return errors.Is(err, slackdomain.ErrNotFound)
}

func validateAssistantPrompt(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return errors.New("assistant prompt is required")
	}
	if len(prompt) > 16<<10 {
		return errors.Join(ErrAssistantPromptTooLarge, errors.New("assistant prompt exceeds 16 KiB"))
	}
	return nil
}

func isPermanentAssistantProviderError(err error) bool {
	var providerError *AssistantAPIError
	return errors.As(err, &providerError) && providerError != nil && providerError.Permanent
}
