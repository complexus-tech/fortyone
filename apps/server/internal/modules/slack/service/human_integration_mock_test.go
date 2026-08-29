package slack

import (
	"context"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
)

func (m *mockRepo) GetSlackWorkspaceForMember(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
) (slackdomain.Installation, error) {
	if err := query.Validate(); err != nil {
		return slackdomain.Installation{}, err
	}
	return m.GetSlackWorkspace(ctx, query.WorkspaceID)
}

func (m *mockRepo) FindSlackUserLinkForMember(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
	slackTeamID string,
) (*slackdomain.UserLink, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return m.FindSlackUserLinkByUser(ctx, query.WorkspaceID, slackTeamID, query.ActorID)
}

func (m *mockRepo) ListChannelsForMember(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
) ([]slackdomain.Channel, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return m.ListChannels(ctx, query.WorkspaceID)
}

func (m *mockRepo) ListRequestLogsForAdmin(
	ctx context.Context,
	query slackdomain.ListRequestLogsQuery,
) ([]slackdomain.RequestLog, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return m.ListRequestLogs(ctx, query.WorkspaceID, int(query.Limit))
}
