package integrationrequests

import (
	"context"
	"testing"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserFacingOperationsRejectActorsWithoutTeamAccess(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	newService := func() (*Service, *requestRepoStub) {
		repo := &requestRepoStub{
			statusID:    uuid.New(),
			deniedUsers: map[uuid.UUID]struct{}{actorID: {}},
			requests: []CoreIntegrationRequest{{
				ID:               requestID,
				WorkspaceID:      workspaceID,
				TeamID:           teamID,
				Provider:         ProviderSlack,
				SourceType:       SourceTypeIssue,
				SourceExternalID: "1",
				Title:            "Private team request",
				Status:           StatusPending,
			}},
		}
		return New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{
			ProviderSlack: providerAccepterStub{},
		}), repo
	}

	tests := []struct {
		name string
		run  func(*Service) error
	}{
		{
			name: "list",
			run: func(service *Service) error {
				_, err := service.ListByTeam(context.Background(), workspaceID, teamID, actorID, CoreListRequestsFilter{})
				return err
			},
		},
		{
			name: "count",
			run: func(service *Service) error {
				_, err := service.CountByTeam(context.Background(), workspaceID, teamID, actorID, CoreListRequestsFilter{})
				return err
			},
		},
		{
			name: "get",
			run: func(service *Service) error {
				_, err := service.GetForUser(context.Background(), workspaceID, requestID, actorID)
				return err
			},
		},
		{
			name: "update",
			run: func(service *Service) error {
				_, err := service.UpdatePending(context.Background(), workspaceID, requestID, actorID, CoreUpdateRequestInput{})
				return err
			},
		},
		{
			name: "accept",
			run: func(service *Service) error {
				_, err := service.Accept(context.Background(), workspaceID, requestID, actorID)
				return err
			},
		},
		{
			name: "decline",
			run: func(service *Service) error {
				_, err := service.Decline(context.Background(), workspaceID, requestID, actorID)
				return err
			},
		},
		{
			name: "accept all",
			run: func(service *Service) error {
				_, err := service.AcceptAllPendingByTeam(context.Background(), workspaceID, teamID, actorID)
				return err
			},
		},
		{
			name: "decline all",
			run: func(service *Service) error {
				_, err := service.DeclineAllPendingByTeam(context.Background(), workspaceID, teamID, actorID)
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, repo := newService()
			err := test.run(service)

			require.ErrorIs(t, err, integrationrequestdomain.ErrNotFound)
			require.Empty(t, repo.createdStories)
			require.Empty(t, repo.markedAccepted)
			require.Empty(t, repo.markedDeclined)
		})
	}
}

func TestCreateCommentPersistsBeforeImmediateProviderDelivery(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	authorID := uuid.New()
	thread := CoreProviderThread{
		ID: uuid.New(), IntegrationRequestID: requestID, WorkspaceID: workspaceID,
		Provider: ProviderSlack, TeamID: teamID,
	}
	comment := CoreIntegrationRequestComment{
		ID: uuid.New(), ThreadID: thread.ID, Direction: CommentDirectionOutbound,
		AuthorUserID: &authorID, Body: "Ship this update to Slack",
	}
	repo := &requestRepoStub{
		requests: []CoreIntegrationRequest{{
			ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
			Provider: ProviderSlack, Status: StatusAccepted,
		}},
		commentThread: thread,
		commentResult: comment,
	}
	commenter := &providerCommenterStub{}
	service := New(nil, repo, storyServiceStub{repo: repo}, nil,
		WithProviderCommenter(ProviderSlack, commenter),
	)

	created, err := service.CreateComment(context.Background(), CoreCreateCommentInput{
		WorkspaceID:          workspaceID,
		RequestID:            requestID,
		AuthorID:             authorID,
		ClientIdempotencyKey: uuid.New(),
		Body:                 "  Ship this update to Slack  ",
	})

	require.NoError(t, err)
	require.Equal(t, comment.ID, created.ID)
	require.Equal(t, "Ship this update to Slack", repo.commentInput.Body)
	require.Equal(t, 1, commenter.calls)
	require.Equal(t, 1, commenter.prepareCalls)
	require.Equal(t, requestID, commenter.request.ID)
	require.Equal(t, thread.ID, commenter.thread.ID)
	require.Equal(t, comment.ID, commenter.comment.ID)
}

func TestCreateCommentReturnsDurableTerminalDuplicateWithoutAnotherProviderSend(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	authorID := uuid.New()
	deliveryStatus := "sent"
	thread := CoreProviderThread{
		ID: uuid.New(), IntegrationRequestID: requestID, WorkspaceID: workspaceID,
		Provider: ProviderSlack, TeamID: teamID,
	}
	comment := CoreIntegrationRequestComment{
		ID: uuid.New(), ThreadID: thread.ID, Direction: CommentDirectionOutbound,
		AuthorUserID: &authorID, DeliveryStatus: &deliveryStatus, Body: "Already delivered",
	}
	repo := &requestRepoStub{
		requests: []CoreIntegrationRequest{{
			ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
			Provider: ProviderSlack, Status: StatusAccepted,
		}},
		commentThread: thread,
		commentResult: comment,
	}
	commenter := &providerCommenterStub{}
	service := New(nil, repo, storyServiceStub{repo: repo}, nil,
		WithProviderCommenter(ProviderSlack, commenter),
	)

	created, err := service.CreateComment(context.Background(), CoreCreateCommentInput{
		WorkspaceID:          workspaceID,
		RequestID:            requestID,
		AuthorID:             authorID,
		ClientIdempotencyKey: uuid.New(),
		Body:                 "Already delivered",
	})

	require.NoError(t, err)
	require.Equal(t, comment.ID, created.ID)
	require.Equal(t, 1, commenter.prepareCalls)
	require.Zero(t, commenter.calls)
}

func TestCreateCommentRequiresCallerIdempotencyKey(t *testing.T) {
	repo := &requestRepoStub{}
	service := New(nil, repo, storyServiceStub{repo: repo}, nil)

	_, err := service.CreateComment(context.Background(), CoreCreateCommentInput{
		WorkspaceID: uuid.New(), RequestID: uuid.New(), AuthorID: uuid.New(), Body: "Post once",
	})

	require.ErrorIs(t, err, ErrInvalidRequestProperty)
	require.ErrorContains(t, err, "comment idempotency key is required")
	require.Equal(t, CoreCreateCommentInput{}, repo.commentInput)
}

func TestCreateCommentRejectsWhitespaceBodyBeforePersistence(t *testing.T) {
	repo := &requestRepoStub{}
	service := New(nil, repo, storyServiceStub{repo: repo}, nil)

	_, err := service.CreateComment(context.Background(), CoreCreateCommentInput{
		WorkspaceID: uuid.New(), RequestID: uuid.New(), AuthorID: uuid.New(), Body: "  ",
	})

	require.ErrorIs(t, err, ErrInvalidRequestProperty)
	require.ErrorContains(t, err, "comment body is required")
	require.Equal(t, CoreCreateCommentInput{}, repo.commentInput)
}
