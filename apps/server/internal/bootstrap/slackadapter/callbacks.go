package slackadapter

import (
	"context"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
)

// ProviderCallbacks translates integration-request callbacks at composition
// time so neither business module imports the other's service models.
type ProviderCallbacks struct {
	service *slack.Service
}

func NewProviderCallbacks(service *slack.Service) *ProviderCallbacks {
	if service == nil {
		return nil
	}
	return &ProviderCallbacks{service: service}
}

func (adapter *ProviderCallbacks) AcceptIntegrationRequest(
	ctx context.Context,
	request integrationrequests.CoreIntegrationRequest,
	story integrationrequests.Story,
) error {
	return adapter.service.AcceptIntegrationRequest(ctx, mapIntegrationRequest(request), slack.Story{
		ID: story.ID, SequenceID: story.SequenceID, Team: story.TeamID, TeamCode: story.TeamCode,
		Title: story.Title, Description: story.Description, Status: story.StatusID,
		Priority: story.Priority, Assignee: story.AssigneeID, Reporter: story.ReporterID,
		EndDate: story.EndDate, CreatedAt: story.CreatedAt, UpdatedAt: story.UpdatedAt,
	})
}

func (adapter *ProviderCallbacks) PrepareIntegrationRequestComment(
	ctx context.Context,
	request integrationrequests.CoreIntegrationRequest,
	thread integrationrequests.CoreProviderThread,
	input integrationrequests.CoreCreateCommentInput,
) (integrationrequests.CorePreparedProviderComment, error) {
	prepared, err := adapter.service.PrepareIntegrationRequestComment(
		ctx,
		mapIntegrationRequest(request),
		mapProviderThread(thread),
		slack.CreateIntegrationRequestCommentInput{
			WorkspaceID: input.WorkspaceID, RequestID: input.RequestID, AuthorID: input.AuthorID,
			ClientIdempotencyKey: input.ClientIdempotencyKey, Body: input.Body,
		},
	)
	return integrationrequests.CorePreparedProviderComment{
		ExternalRecipientUserID: prepared.ExternalRecipientUserID,
		ProviderPayload:         append([]byte(nil), prepared.ProviderPayload...),
	}, err
}

func (adapter *ProviderCallbacks) DeliverIntegrationRequestComment(
	ctx context.Context,
	request integrationrequests.CoreIntegrationRequest,
	thread integrationrequests.CoreProviderThread,
	comment integrationrequests.CoreIntegrationRequestComment,
	prepared integrationrequests.CorePreparedProviderComment,
) error {
	return adapter.service.DeliverIntegrationRequestComment(
		ctx,
		mapIntegrationRequest(request),
		mapProviderThread(thread),
		slack.IntegrationRequestComment{
			ID: comment.ID, WorkspaceID: comment.WorkspaceID, ThreadID: comment.ThreadID,
			Direction: comment.Direction, AuthorUserID: comment.AuthorUserID,
			AuthorName: comment.AuthorName, AuthorAvatar: comment.AuthorAvatar,
			ExternalAuthorID: comment.ExternalAuthorID, ExternalMessageID: comment.ExternalMessageID,
			ClientIdempotencyKey:   comment.ClientIdempotencyKey,
			OutboundIdempotencyKey: comment.OutboundIdempotencyKey,
			DeliveryStatus:         comment.DeliveryStatus, Body: comment.Body,
			CreatedAt: comment.CreatedAt, UpdatedAt: comment.UpdatedAt,
		},
		slack.PreparedProviderComment{
			ExternalRecipientUserID: prepared.ExternalRecipientUserID,
			ProviderPayload:         append([]byte(nil), prepared.ProviderPayload...),
		},
	)
}

var (
	_ integrationrequests.ProviderAccepter  = (*ProviderCallbacks)(nil)
	_ integrationrequests.ProviderCommenter = (*ProviderCallbacks)(nil)
)
