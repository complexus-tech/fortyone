package slack

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// SlackStoryReader is deliberately read-only. Work Object events never mutate
// a story and never bypass the actor/channel authorization checks below.
type SlackStoryReader interface {
	QueryByRef(ctx context.Context, workspaceID uuid.UUID, storyRef string) (singleStory, error)
}

// SlackRequestReader is deliberately permission-aware and read-only. Its
// contract requires the current FortyOne actor so request previews cannot
// bypass workspace or team membership.
type SlackRequestReader interface {
	GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (integrationRequest, error)
}

// SlackObjectiveReader resolves one objective through the actor-aware path so
// Slack previews cannot bypass objective visibility rules.
type SlackObjectiveReader interface {
	ListByID(ctx context.Context, workspaceID, userID, objectiveID uuid.UUID) ([]objective, error)
}

// SlackSprintReader resolves one sprint through the actor-aware path so Slack
// previews cannot bypass team membership checks.
type SlackSprintReader interface {
	ListByID(ctx context.Context, workspaceID, userID, sprintID uuid.UUID) ([]sprint, error)
}

type slackWorkObjectRepository interface {
	ListAuthorizedChannelTeamIDs(
		ctx context.Context,
		workspaceID, slackWorkspaceID uuid.UUID,
		slackChannelID string,
		userID uuid.UUID,
	) ([]uuid.UUID, error)
	ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]slackStatusRecord, error)
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]slackTeamMemberRecord, error)
	FindTeamMemberByID(ctx context.Context, teamID, userID uuid.UUID) (slackTeamMemberRecord, error)
}

func isSlackWorkObjectEvent(kind slackEventKind) bool {
	return kind == slackEventKindLinkShared
}

func (p *EventProcessor) processSlackWorkObjectEvent(
	ctx context.Context,
	workspace workspaceRecord,
	installation slackWorkspaceRecord,
	linkedUserID *uuid.UUID,
	event normalizedSlackEvent,
	botToken string,
) error {
	if (p.storyReader == nil && p.requestReader == nil && p.objectiveReader == nil && p.sprintReader == nil) || p.workObjects == nil {
		return errors.New("slack Work Object runtime is not configured")
	}
	switch event.Kind {
	case slackEventKindLinkShared:
		return p.processSlackLinkShared(ctx, workspace, installation, linkedUserID, event, botToken)
	default:
		return nil
	}
}
