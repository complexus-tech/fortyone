package slack

import (
	"context"

	"github.com/google/uuid"
)

func (p *EventProcessor) routeAssistantThreadEvent(
	ctx context.Context,
	workspaceID uuid.UUID,
	installation slackWorkspaceRecord,
	linkedUserID *uuid.UUID,
	event normalizedSlackEvent,
) (status string, done bool, err error) {
	requestThreadAssistantSubscribed := false
	if p.threadSync != nil {
		handled, syncErr := p.syncIntegrationRequestThreadReply(ctx, installation, linkedUserID, event)
		if syncErr != nil {
			return "failed", true, syncErr
		}
		if handled && event.Kind != slackEventKindMention {
			broadMentionDuplicate := event.Kind == slackEventKindChannelThread &&
				installation.BotUserID != nil &&
				containsSlackUserMention(event.Text, *installation.BotUserID)
			if !broadMentionDuplicate {
				if linkedUserID == nil || *linkedUserID == uuid.Nil {
					return "completed", true, nil
				}
				requestThreadAssistantSubscribed, syncErr = p.channelThreadIsSubscribed(ctx, workspaceID, *linkedUserID, installation, event)
				if syncErr != nil {
					return "failed", true, syncErr
				}
				if !requestThreadAssistantSubscribed {
					return "completed", true, nil
				}
			}
		}
	}
	if event.Kind != slackEventKindChannelThread {
		return "", false, nil
	}

	// An explicit app_mention event owns messages that mention the bot. Slack
	// can also emit the same Slack message through message.channels/groups;
	// ignoring that broad event prevents two answers with different event IDs.
	if installation.BotUserID != nil && containsSlackUserMention(event.Text, *installation.BotUserID) {
		return "ignored", true, nil
	}
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		return "ignored", true, nil
	}
	if requestThreadAssistantSubscribed {
		return "", false, nil
	}
	subscribed, subscriptionErr := p.channelThreadIsSubscribed(ctx, workspaceID, *linkedUserID, installation, event)
	if subscriptionErr != nil {
		return "failed", true, subscriptionErr
	}
	if !subscribed {
		return "ignored", true, nil
	}
	return "", false, nil
}
