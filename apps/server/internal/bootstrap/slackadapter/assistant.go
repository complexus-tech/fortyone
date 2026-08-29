package slackadapter

import (
	"context"
	"errors"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/google/uuid"
)

type Assistant struct {
	backend messaging.Assistant
}

func NewAssistant(backend messaging.Assistant) *Assistant {
	if backend == nil {
		return nil
	}
	return &Assistant{backend: backend}
}

func (adapter *Assistant) Respond(ctx context.Context, input slack.AssistantRequest) (slack.AssistantResponse, error) {
	response, err := adapter.backend.Respond(ctx, messaging.Request{
		WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		AllowedTeamIDs: append([]uuid.UUID(nil), input.AllowedTeamIDs...),
		SharedTeamIDs:  append([]uuid.UUID(nil), input.SharedTeamIDs...),
		RuntimeContext: mapRuntimeContextToMessaging(input.RuntimeContext),
		Guidance:       input.Guidance, AllowMutations: input.AllowMutations,
		WebsiteURL: input.WebsiteURL, SourceURL: input.SourceURL,
		Conversation: mapConversationToMessaging(input.Conversation), Prompt: input.Prompt,
	})
	return mapAssistantResponse(response), mapAssistantError(err)
}

func mapConversationToMessaging(turns []slack.AssistantConversationTurn) []messaging.ConversationTurn {
	if turns == nil {
		return nil
	}
	result := make([]messaging.ConversationTurn, 0, len(turns))
	for _, turn := range turns {
		result = append(result, messaging.ConversationTurn{
			Role: messaging.ConversationRole(turn.Role),
			Text: turn.Text,
		})
	}
	return result
}

func mapRuntimeContextToMessaging(input *slack.AssistantRuntimeContext) *messaging.RuntimeContext {
	if input == nil {
		return nil
	}
	result := &messaging.RuntimeContext{
		Actor: messaging.RuntimeActorContext{DisplayName: input.Actor.DisplayName, Username: input.Actor.Username},
		Workspace: messaging.RuntimeWorkspaceContext{
			Name: input.Workspace.Name, Slug: input.Workspace.Slug, Role: input.Workspace.Role,
		},
		LocalTime: input.LocalTime,
		Terminology: messaging.RuntimeTerminologyContext{
			Story:     messaging.RuntimeTerm{Singular: input.Terminology.Story.Singular, Plural: input.Terminology.Story.Plural},
			Sprint:    messaging.RuntimeTerm{Singular: input.Terminology.Sprint.Singular, Plural: input.Terminology.Sprint.Plural},
			Objective: messaging.RuntimeTerm{Singular: input.Terminology.Objective.Singular, Plural: input.Terminology.Objective.Plural},
			KeyResult: messaging.RuntimeTerm{Singular: input.Terminology.KeyResult.Singular, Plural: input.Terminology.KeyResult.Plural},
		},
		Surface: mapSurfaceToMessaging(input.Surface),
	}
	if input.TeamHints != nil {
		result.TeamHints = make([]messaging.RuntimeTeamHint, 0, len(input.TeamHints))
		for _, hint := range input.TeamHints {
			result.TeamHints = append(result.TeamHints, messaging.RuntimeTeamHint{Name: hint.Name, Code: hint.Code})
		}
	}
	return result
}

func mapSurfaceToMessaging(input slack.AssistantSurfaceContext) messaging.RuntimeSurfaceContext {
	result := messaging.RuntimeSurfaceContext{
		Provider: input.Provider, Kind: messaging.RuntimeSurfaceKind(input.Kind), Location: input.Location,
	}
	if input.CurrentEntity != nil {
		result.CurrentEntity = &messaging.RuntimeEntityHint{
			Kind: input.CurrentEntity.Kind, Reference: input.CurrentEntity.Reference, Title: input.CurrentEntity.Title,
		}
	}
	return result
}

func mapAssistantResponse(response messaging.Response) slack.AssistantResponse {
	result := slack.AssistantResponse{
		Text: response.Text,
		Usage: slack.AssistantUsage{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
			TotalTokens:  response.Usage.TotalTokens,
		},
	}
	if response.Confirmation != nil {
		result.Confirmation = &slack.StoryMutationConfirmation{
			Operation: slack.StoryMutationOperation(response.Confirmation.Operation),
			Token:     response.Confirmation.Token, ExpiresAt: response.Confirmation.ExpiresAt,
			Prompt: response.Confirmation.Prompt,
		}
	}
	return result
}

func mapAssistantError(err error) error {
	switch {
	case errors.Is(err, messaging.ErrAssistantNotConfigured):
		return errors.Join(slack.ErrAssistantNotConfigured, err)
	case errors.Is(err, messaging.ErrMessageTooLarge):
		return errors.Join(slack.ErrAssistantPromptTooLarge, err)
	}
	var apiError *messaging.APIError
	if errors.As(err, &apiError) && apiError != nil {
		return &slack.AssistantAPIError{
			StatusCode: apiError.StatusCode, Code: apiError.Code, Message: apiError.Error(),
			RequestID: apiError.RequestID, Permanent: messaging.IsPermanentOpenAIError(err),
		}
	}
	return err
}

type ContextBackend interface {
	Load(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID, messaging.RuntimeSurfaceContext, time.Time) (*messaging.RuntimeContext, error)
}

type ContextProvider struct {
	backend ContextBackend
}

func NewContextProvider(backend ContextBackend) *ContextProvider {
	if backend == nil {
		return nil
	}
	return &ContextProvider{backend: backend}
}

func (adapter *ContextProvider) Load(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	allowedTeamIDs []uuid.UUID,
	surface slack.AssistantSurfaceContext,
	now time.Time,
) (*slack.AssistantRuntimeContext, error) {
	runtime, err := adapter.backend.Load(
		ctx,
		workspaceID,
		userID,
		append([]uuid.UUID(nil), allowedTeamIDs...),
		mapSurfaceToMessaging(surface),
		now,
	)
	if err != nil || runtime == nil {
		return nil, err
	}
	return mapRuntimeContextFromMessaging(runtime), nil
}

func mapRuntimeContextFromMessaging(input *messaging.RuntimeContext) *slack.AssistantRuntimeContext {
	result := &slack.AssistantRuntimeContext{
		Actor: slack.AssistantActorContext{DisplayName: input.Actor.DisplayName, Username: input.Actor.Username},
		Workspace: slack.AssistantWorkspaceContext{
			Name: input.Workspace.Name, Slug: input.Workspace.Slug, Role: input.Workspace.Role,
		},
		LocalTime: input.LocalTime,
		Terminology: slack.AssistantTerminologyContext{
			Story:     slack.AssistantTerm{Singular: input.Terminology.Story.Singular, Plural: input.Terminology.Story.Plural},
			Sprint:    slack.AssistantTerm{Singular: input.Terminology.Sprint.Singular, Plural: input.Terminology.Sprint.Plural},
			Objective: slack.AssistantTerm{Singular: input.Terminology.Objective.Singular, Plural: input.Terminology.Objective.Plural},
			KeyResult: slack.AssistantTerm{Singular: input.Terminology.KeyResult.Singular, Plural: input.Terminology.KeyResult.Plural},
		},
		Surface: mapSurfaceFromMessaging(input.Surface),
	}
	if input.TeamHints != nil {
		result.TeamHints = make([]slack.AssistantTeamHint, 0, len(input.TeamHints))
		for _, hint := range input.TeamHints {
			result.TeamHints = append(result.TeamHints, slack.AssistantTeamHint{Name: hint.Name, Code: hint.Code})
		}
	}
	return result
}

func mapSurfaceFromMessaging(input messaging.RuntimeSurfaceContext) slack.AssistantSurfaceContext {
	result := slack.AssistantSurfaceContext{
		Provider: input.Provider, Kind: slack.AssistantSurfaceKind(input.Kind), Location: input.Location,
	}
	if input.CurrentEntity != nil {
		result.CurrentEntity = &slack.AssistantEntityHint{
			Kind: input.CurrentEntity.Kind, Reference: input.CurrentEntity.Reference, Title: input.CurrentEntity.Title,
		}
	}
	return result
}

var (
	_ slack.Assistant                = (*Assistant)(nil)
	_ slack.AssistantContextProvider = (*ContextProvider)(nil)
)
