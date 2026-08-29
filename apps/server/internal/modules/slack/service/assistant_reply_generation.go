package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (p *EventProcessor) prepareAssistantReply(
	ctx context.Context,
	input assistantEventInput,
	state *assistantDeliveryState,
) error {
	if state.reply == "" {
		if _, err := p.usageBudget.Check(ctx, input.workspace.ID, p.dailyWorkspaceTokenLimit); err != nil {
			if !errors.Is(err, errDailyWorkspaceTokenLimit) {
				if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(err)); failErr != nil {
					return errors.Join(err, failErr)
				}
				return err
			}
			state.reply = assistantDailyLimitReply
			state.persistConversation = false
		}
	}
	if state.reply == "" {
		admission, admissionErr := p.callLimiter.Admit(ctx, AssistantAdmissionInput{
			Provider:            "slack",
			WorkspaceID:         input.workspace.ID,
			UserID:              input.linkedUserID,
			ExternalWorkspaceID: input.event.TeamID,
			ExternalEventID:     input.event.EventID,
		})
		if admissionErr != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(admissionErr)); failErr != nil {
				return errors.Join(admissionErr, failErr)
			}
			return admissionErr
		}
		if !admission.Allowed {
			state.reply = assistantWorkspaceRateReply
			if admission.LimitedScope == "user" {
				state.reply = assistantUserRateLimitReply
			}
			state.persistConversation = false
		}
	}
	if input.event.Kind != slackEventKindDirect && isAssistantBudgetNotice(state.reply) && state.channelID != input.event.UserID {
		if err := p.store.SetOutboundDeliveryContentAndDestination(ctx, state.record.ID, state.reply, input.event.UserID, ""); err != nil {
			return err
		}
		state.channelID = input.event.UserID
		state.threadTS = ""
		state.contentPersisted = true
	}
	if state.persistConversation {
		conversationID, err := p.persistAssistantPrompt(
			ctx,
			input.workspace.ID,
			input.linkedUserID,
			input.event,
			input.prompt,
			input.allowedTeamIDs,
			input.sharedTeamIDs,
		)
		if err != nil {
			return err
		}
		state.conversationID = conversationID
	}
	if state.reply != "" {
		return nil
	}

	state.thinkingStatusActive = p.startAssistantThinkingStatus(ctx, input.event, input.botToken)
	return p.generateAssistantReply(ctx, input, state)
}

func (p *EventProcessor) generateAssistantReply(
	ctx context.Context,
	input assistantEventInput,
	state *assistantDeliveryState,
) error {
	history, err := p.store.ListRecentMessages(ctx, state.conversationID, slackConversationHistoryLimit)
	if err != nil {
		return err
	}
	turns := make([]assistantConversationTurn, 0, len(history)+1)
	excludedThreadMessageIDs := map[string]struct{}{
		strings.TrimSpace(input.event.MessageTS): {},
	}
	for _, message := range history {
		if message.ExternalMessageID != nil {
			externalMessageID := strings.TrimSpace(*message.ExternalMessageID)
			if externalMessageID != "" {
				excludedThreadMessageIDs[externalMessageID] = struct{}{}
			}
			if externalMessageID == input.event.MessageTS && message.Role == "user" {
				continue
			}
		}
		role := assistantRoleUser
		if message.Role == "assistant" {
			role = assistantRoleAssistant
		}
		turns = append(turns, assistantConversationTurn{Role: role, Text: message.Content})
	}

	threadSourceURL := slackThreadSourceURL(input.installation, input.event)
	if slackEventCanHydrateThread(input.event) && slackPromptRequestsThreadContext(input.prompt) {
		threadReference, threadErr := p.loadSlackThreadReference(
			ctx,
			input.botToken,
			input.installation,
			input.event,
			excludedThreadMessageIDs,
		)
		if threadErr != nil {
			if slackThreadContextInvalidatesInstallation(threadErr) {
				deactivateErr := p.repo.DeactivateSlackWorkspaceByTeamID(
					ctx,
					input.event.TeamID,
					input.installation.InstallGeneration,
				)
				if deactivateErr != nil && !isSlackRepositoryNotFound(deactivateErr) {
					threadErr = errors.Join(threadErr, fmt.Errorf("deactivate invalid Slack installation: %w", deactivateErr))
				}
			}
			if failureReply, handled := slackThreadContextFailureReply(threadErr); handled {
				state.reply = failureReply
			} else {
				if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(threadErr)); failErr != nil {
					return errors.Join(threadErr, failErr)
				}
				return threadErr
			}
		} else {
			// Keep imported participants out of persisted conversation state.
			// This final ephemeral turn survives the assistant's newest-history
			// bound while remaining clearly marked as untrusted reference data.
			turns = append(turns, threadReference.Turn)
		}
	}
	if state.reply != "" {
		return nil
	}

	runtimeContext, contextErr := p.contextProvider.Load(
		ctx,
		input.workspace.ID,
		input.linkedUserID,
		input.allowedTeamIDs,
		assistantSurfaceForSlackEvent(input.event),
		p.clock.Now(),
	)
	if contextErr != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(contextErr)); failErr != nil {
			return errors.Join(contextErr, failErr)
		}
		return contextErr
	}
	response, responseErr := p.assistant.Respond(ctx, assistantRequest{
		WorkspaceID:    input.workspace.ID,
		UserID:         input.linkedUserID,
		AllowedTeamIDs: input.allowedTeamIDs,
		SharedTeamIDs:  input.sharedTeamIDs,
		RuntimeContext: runtimeContext,
		Guidance:       input.agentSettings.Guidance,
		AllowMutations: true,
		WebsiteURL:     p.website,
		SourceURL:      threadSourceURL,
		Conversation:   turns,
		Prompt:         input.prompt,
	})
	if responseErr != nil {
		p.logAssistantResponseError(
			ctx,
			responseErr,
			input.workspace.ID,
			input.linkedUserID,
			input.receipt.ID,
			input.receipt.AttemptCount,
			input.event,
		)
	}
	_, usageErr := p.recordAssistantUsage(ctx, dailyUsageRecordInput{
		InboundEventID:      input.receipt.ID,
		WorkspaceID:         input.workspace.ID,
		Provider:            "slack",
		ExternalWorkspaceID: input.event.TeamID,
		ExternalEventID:     input.event.EventID,
		AttemptCount:        input.receipt.AttemptCount,
		Usage:               response.Usage,
	})
	if usageErr != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(usageErr)); failErr != nil {
			return errors.Join(responseErr, usageErr, failErr)
		}
		return errors.Join(responseErr, usageErr)
	}
	if responseErr != nil {
		if errors.Is(responseErr, errAssistantNotConfigured) || isPermanentAssistantProviderError(responseErr) {
			state.reply = assistantConfigurationReply
		} else {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(responseErr)); failErr != nil {
				return errors.Join(responseErr, failErr)
			}
			return responseErr
		}
	} else if response.Confirmation != nil {
		state.reply = truncateSlackText(response.Confirmation.Prompt)
		confirmationPayload, confirmationErr := BuildSlackMutationConfirmationProviderPayload(
			state.reply,
			response.Confirmation.Token,
			input.event.UserID,
			response.Confirmation.Operation == storyMutationCreateBatch,
		)
		if confirmationErr != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(confirmationErr)); failErr != nil {
				return errors.Join(confirmationErr, failErr)
			}
			return confirmationErr
		}
		confirmationPayload.Authorization = state.providerPayload.Authorization
		state.providerPayload = confirmationPayload
	} else {
		state.reply = truncateSlackText(response.Text)
	}
	if state.reply == "" {
		state.reply = "I couldn't generate a useful response. Please try again."
	}
	return nil
}
