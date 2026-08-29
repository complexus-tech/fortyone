package providers

import (
	"context"
	"fmt"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

type outboundStoryEventPublisher interface {
	Publish(context.Context, outboundwebhooksdomain.PublishEvent) (int, error)
}

// OutboundStoryMutationPublisher is a bootstrap-owned adapter. Stories know
// only their narrow publication port; this layer translates committed story
// intents into the outbound-webhook product event catalog.
type OutboundStoryMutationPublisher struct {
	publisher outboundStoryEventPublisher
}

func NewOutboundStoryMutationPublisher(publisher outboundStoryEventPublisher) (*OutboundStoryMutationPublisher, error) {
	if publisher == nil {
		return nil, fmt.Errorf("outbound story mutation publisher is required")
	}
	return &OutboundStoryMutationPublisher{publisher: publisher}, nil
}

func (adapter *OutboundStoryMutationPublisher) PublishStoryMutationEvent(
	ctx context.Context,
	event stories.StoryMutationPublication,
) error {
	eventType, err := outboundStoryMutationEventType(event.Type)
	if err != nil {
		return err
	}
	_, err = adapter.publisher.Publish(ctx, outboundwebhooksdomain.PublishEvent{
		ID: event.ID, WorkspaceID: event.WorkspaceID, Type: eventType,
		SubjectID: event.StoryID, Actor: event.Actor,
		Payload: append([]byte(nil), event.Payload...), OccurredAt: event.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("publish story mutation event: %w", err)
	}
	return nil
}

func outboundStoryMutationEventType(value string) (outboundwebhooksdomain.EventType, error) {
	switch value {
	case string(outboundwebhooksdomain.EventStoryCreated):
		return outboundwebhooksdomain.EventStoryCreated, nil
	case string(outboundwebhooksdomain.EventStoryUpdated):
		return outboundwebhooksdomain.EventStoryUpdated, nil
	case string(outboundwebhooksdomain.EventStoryDeleted):
		return outboundwebhooksdomain.EventStoryDeleted, nil
	default:
		return "", fmt.Errorf("unsupported story mutation event type %q", value)
	}
}
