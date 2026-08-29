package integrationrequests

import (
	"context"
	"errors"
	"testing"
	"time"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type requestRepoStub struct {
	requests        []CoreIntegrationRequest
	markedAccepted  []uuid.UUID
	markedDeclined  []uuid.UUID
	statusID        uuid.UUID
	createdStories  []NewStory
	deniedUsers     map[uuid.UUID]struct{}
	commentInput    CoreCreateCommentInput
	commentPrepared CorePreparedProviderComment
	commentThread   CoreProviderThread
	commentResult   CoreIntegrationRequestComment
	reserveCalls    int
}

func (r *requestRepoStub) UpsertPending(ctx context.Context, input CoreUpsertRequestInput) (CoreIntegrationRequest, error) {
	return CoreIntegrationRequest{}, nil
}

func (r *requestRepoStub) AuthorizeTeam(_ context.Context, _, _ uuid.UUID, userID uuid.UUID) error {
	if _, denied := r.deniedUsers[userID]; denied {
		return integrationrequestdomain.ErrNotFound
	}
	return nil
}

func (r *requestRepoStub) ListByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter CoreListRequestsFilter) ([]CoreIntegrationRequest, error) {
	if err := r.AuthorizeTeam(ctx, workspaceID, teamID, userID); err != nil {
		return nil, err
	}
	result := make([]CoreIntegrationRequest, 0, len(r.requests))
	for _, request := range r.requests {
		if request.WorkspaceID != workspaceID || request.TeamID != teamID {
			continue
		}
		if filter.Status != "" && request.Status != filter.Status {
			continue
		}
		if filter.Provider != "" && request.Provider != filter.Provider {
			continue
		}
		if filter.Priority != "" && request.Priority != filter.Priority {
			continue
		}
		if filter.AssigneeID != nil && (request.AssigneeID == nil || *request.AssigneeID != *filter.AssigneeID) {
			continue
		}
		if filter.CreatedAfter != nil && request.CreatedAt.Before(*filter.CreatedAfter) {
			continue
		}
		if filter.CreatedBefore != nil && request.CreatedAt.After(*filter.CreatedBefore) {
			continue
		}
		result = append(result, request)
	}
	if filter.PageSize > 0 {
		page := filter.Page
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * filter.PageSize
		if offset >= len(result) {
			return []CoreIntegrationRequest{}, nil
		}
		end := offset + filter.PageSize
		if end > len(result) {
			end = len(result)
		}
		result = result[offset:end]
	}
	return result, nil
}

func (r *requestRepoStub) CountByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter CoreListRequestsFilter) (int, error) {
	filter.Page = 0
	filter.PageSize = 0
	requests, err := r.ListByTeam(ctx, workspaceID, teamID, userID, filter)
	return len(requests), err
}

func (r *requestRepoStub) GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (CoreIntegrationRequest, error) {
	for _, request := range r.requests {
		if request.WorkspaceID == workspaceID && request.ID == requestID {
			if err := r.AuthorizeTeam(ctx, workspaceID, request.TeamID, userID); err != nil {
				return CoreIntegrationRequest{}, err
			}
			return request, nil
		}
	}
	return CoreIntegrationRequest{}, integrationrequestdomain.ErrNotFound
}

func (r *requestRepoStub) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	return &r.statusID, nil
}

func (r *requestRepoStub) UpdatePending(ctx context.Context, workspaceID, requestID, userID uuid.UUID, input CoreUpdateRequestInput) (CoreIntegrationRequest, error) {
	return r.GetForUser(ctx, workspaceID, requestID, userID)
}

func (r *requestRepoStub) ReserveAcceptance(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (CoreIntegrationRequest, error) {
	r.reserveCalls++
	request, err := r.GetForUser(ctx, workspaceID, requestID, userID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	if request.Status != StatusPending {
		return CoreIntegrationRequest{}, integrationrequestdomain.ErrNotFound
	}
	if request.AcceptanceState == "" || request.AcceptanceState == AcceptanceStateIdle {
		request.AcceptanceState = AcceptanceStateReserved
		request.AcceptanceStartedByUserID = &userID
		for index := range r.requests {
			if r.requests[index].ID == requestID {
				r.requests[index] = request
				break
			}
		}
	}
	return request, nil
}

func (r *requestRepoStub) MarkAccepted(_ context.Context, workspaceID, requestID, storyID, acceptedByUserID uuid.UUID) (CoreIntegrationRequest, error) {
	var request CoreIntegrationRequest
	found := false
	for _, candidate := range r.requests {
		if candidate.WorkspaceID == workspaceID && candidate.ID == requestID {
			request = candidate
			found = true
			break
		}
	}
	if !found || request.Status != StatusPending || request.AcceptanceState != AcceptanceStateReserved || request.AcceptanceStartedByUserID == nil || *request.AcceptanceStartedByUserID != acceptedByUserID {
		return CoreIntegrationRequest{}, integrationrequestdomain.ErrNotFound
	}
	r.markedAccepted = append(r.markedAccepted, requestID)
	request.Status = StatusAccepted
	request.AcceptedStoryID = &storyID
	request.AcceptedByUserID = &acceptedByUserID
	request.AcceptanceState = AcceptanceStateIdle
	request.AcceptanceStartedByUserID = nil
	for index := range r.requests {
		if r.requests[index].ID == requestID {
			r.requests[index] = request
			break
		}
	}
	return request, nil
}

func (r *requestRepoStub) MarkDeclined(ctx context.Context, workspaceID, requestID, declinedByUserID uuid.UUID) (CoreIntegrationRequest, error) {
	r.markedDeclined = append(r.markedDeclined, requestID)
	request, err := r.GetForUser(ctx, workspaceID, requestID, declinedByUserID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	if request.AcceptanceState == AcceptanceStateReserved {
		return CoreIntegrationRequest{}, integrationrequestdomain.ErrNotFound
	}
	request.Status = StatusDeclined
	for index := range r.requests {
		if r.requests[index].ID == requestID {
			r.requests[index] = request
			break
		}
	}
	return request, nil
}

func (r *requestRepoStub) BindProviderThread(context.Context, CoreBindProviderThreadInput) (CoreProviderThread, error) {
	return CoreProviderThread{}, nil
}

func (r *requestRepoStub) HasAuthorizedProviderThread(context.Context, CoreProviderThreadMatchInput) (bool, error) {
	return false, nil
}

func (r *requestRepoStub) HasCurrentProviderThread(context.Context, CoreProviderThreadLookupInput) (bool, error) {
	return false, nil
}

func (r *requestRepoStub) FindProviderThread(context.Context, uuid.UUID, uuid.UUID, string) (CoreProviderThread, error) {
	if r.commentThread.ID == uuid.Nil {
		return CoreProviderThread{}, ErrProviderThreadNotFound
	}
	return r.commentThread, nil
}

func (r *requestRepoStub) GetThreadActivityForRequest(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreThreadActivity, error) {
	return CoreThreadActivity{}, nil
}

func (r *requestRepoStub) ListProviderThreadsForStory(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]CoreProviderThread, error) {
	return nil, nil
}

func (r *requestRepoStub) CreateOutboundComment(_ context.Context, input CoreCreateCommentInput, prepared CorePreparedProviderComment) (CoreProviderThread, CoreIntegrationRequestComment, error) {
	r.commentInput = input
	r.commentPrepared = prepared
	return r.commentThread, r.commentResult, nil
}

func (r *requestRepoStub) GetCommentForUser(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreIntegrationRequestComment, error) {
	return r.commentResult, nil
}

type storyServiceStub struct {
	repo *requestRepoStub
}

type idempotentStoryServiceStub struct {
	calls   int
	stories map[string]Story
}

func (s *idempotentStoryServiceStub) CreateForIntegrationRequest(_ context.Context, _, _ uuid.UUID, input NewStory) (Story, error) {
	s.calls++
	if input.CreationKey == "" {
		return Story{}, errors.New("creation key is required")
	}
	if existing, ok := s.stories[input.CreationKey]; ok {
		return existing, nil
	}
	created := Story{ID: uuid.New()}
	s.stories[input.CreationKey] = created
	return created, nil
}

func (s storyServiceStub) CreateForIntegrationRequest(_ context.Context, _, _ uuid.UUID, input NewStory) (Story, error) {
	s.repo.createdStories = append(s.repo.createdStories, input)
	return Story{ID: uuid.New()}, nil
}

type providerAccepterStub struct{}

func (providerAccepterStub) AcceptIntegrationRequest(context.Context, CoreIntegrationRequest, Story) error {
	return nil
}

type flakyProviderAccepterStub struct {
	calls        int
	failuresLeft int
	storyIDs     []uuid.UUID
}

func (s *flakyProviderAccepterStub) AcceptIntegrationRequest(_ context.Context, _ CoreIntegrationRequest, story Story) error {
	s.calls++
	s.storyIDs = append(s.storyIDs, story.ID)
	if s.failuresLeft > 0 {
		s.failuresLeft--
		return errors.New("provider delivery failed")
	}
	return nil
}

type providerCommenterStub struct {
	prepareCalls int
	calls        int
	request      CoreIntegrationRequest
	thread       CoreProviderThread
	comment      CoreIntegrationRequestComment
	prepared     CorePreparedProviderComment
}

func (s *providerCommenterStub) PrepareIntegrationRequestComment(_ context.Context, _ CoreIntegrationRequest, _ CoreProviderThread, _ CoreCreateCommentInput) (CorePreparedProviderComment, error) {
	s.prepareCalls++
	return s.prepared, nil
}

func (s *providerCommenterStub) DeliverIntegrationRequestComment(_ context.Context, request CoreIntegrationRequest, thread CoreProviderThread, comment CoreIntegrationRequestComment, _ CorePreparedProviderComment) error {
	s.calls++
	s.request = request
	s.thread = thread
	s.comment = comment
	return nil
}

func TestAcceptAllPendingByTeamAcceptsEveryPendingRequest(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	firstID := uuid.New()
	secondID := uuid.New()
	repo := &requestRepoStub{
		statusID: uuid.New(),
		requests: []CoreIntegrationRequest{
			{ID: firstID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "1", Title: "First", Priority: "High", Status: StatusPending},
			{ID: secondID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderSlack, SourceType: SourceTypeIssue, SourceExternalID: "2", Title: "Second", Status: StatusPending},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: uuid.New(), Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "3", Title: "Other team", Status: StatusPending},
		},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{
		ProviderGitHub: providerAccepterStub{},
		ProviderSlack:  providerAccepterStub{},
	})

	result, err := service.AcceptAllPendingByTeam(context.Background(), workspaceID, teamID, actorID)

	require.NoError(t, err)
	require.Equal(t, 2, result.Count)
	require.Equal(t, 2, result.TotalCount)
	require.Equal(t, 2, result.SucceededCount)
	require.Zero(t, result.FailedCount)
	require.False(t, result.Partial)
	require.Len(t, result.Items, 2)
	require.ElementsMatch(t, []uuid.UUID{firstID, secondID}, result.RequestIDs)
	require.ElementsMatch(t, []uuid.UUID{firstID, secondID}, repo.markedAccepted)
	require.Len(t, repo.createdStories, 2)
	require.Equal(t, "High", repo.createdStories[0].Priority)
	require.Equal(t, "No Priority", repo.createdStories[1].Priority)
	for index, created := range repo.createdStories {
		require.NotEmpty(t, created.CreationKey, "story %d has no idempotency key", index)
		require.Contains(t, created.CreationKey, workspaceID.String())
	}
}

func TestAcceptAllPendingByTeamContinuesAndReportsPerItemFailures(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	firstID := uuid.New()
	failedID := uuid.New()
	thirdID := uuid.New()
	repo := &requestRepoStub{
		statusID: uuid.New(),
		requests: []CoreIntegrationRequest{
			{ID: firstID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "1", Title: "First", Status: StatusPending},
			{ID: failedID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderIntercom, SourceType: SourceTypeIssue, SourceExternalID: "2", Title: "Unsupported", Status: StatusPending},
			{ID: thirdID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderSlack, SourceType: SourceTypeIssue, SourceExternalID: "3", Title: "Third", Status: StatusPending},
		},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{
		ProviderGitHub: providerAccepterStub{},
		ProviderSlack:  providerAccepterStub{},
	})

	result, err := service.AcceptAllPendingByTeam(context.Background(), workspaceID, teamID, actorID)

	require.NoError(t, err)
	require.Equal(t, 3, result.TotalCount)
	require.Equal(t, 2, result.SucceededCount)
	require.Equal(t, 1, result.FailedCount)
	require.True(t, result.Partial)
	require.Equal(t, []uuid.UUID{firstID, thirdID}, result.RequestIDs)
	require.Len(t, result.Items, 3)
	require.Equal(t, firstID, result.Items[0].RequestID)
	require.True(t, result.Items[0].Success)
	require.NotNil(t, result.Items[0].AcceptedStoryID)
	require.Equal(t, failedID, result.Items[1].RequestID)
	require.False(t, result.Items[1].Success)
	require.ErrorContains(t, errors.New(result.Items[1].Error), ErrUnsupportedProvider.Error())
	require.Equal(t, thirdID, result.Items[2].RequestID)
	require.True(t, result.Items[2].Success)
}

func TestAcceptPreservesValidatedSlackRequestLabels(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	requestID := uuid.New()
	firstLabelID := uuid.New()
	secondLabelID := uuid.New()
	repo := &requestRepoStub{
		statusID: uuid.New(),
		requests: []CoreIntegrationRequest{{
			ID:               requestID,
			WorkspaceID:      workspaceID,
			TeamID:           teamID,
			Provider:         ProviderSlack,
			SourceType:       "message",
			SourceExternalID: "Ev1",
			Title:            "Create from Slack",
			Status:           StatusPending,
			LabelIDs:         []uuid.UUID{firstLabelID, secondLabelID},
		}},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{
		ProviderSlack: providerAccepterStub{},
	})

	_, err := service.Accept(context.Background(), workspaceID, requestID, actorID)

	require.NoError(t, err)
	require.Len(t, repo.createdStories, 1)
	require.Equal(t, []uuid.UUID{firstLabelID, secondLabelID}, repo.createdStories[0].LabelIDs)
	require.Equal(t, "integration-request:"+workspaceID.String()+":"+requestID.String(), repo.createdStories[0].CreationKey)
}

func TestAcceptResumesReservedConversionWithoutCreatingAnotherStory(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	retryActorID := uuid.New()
	requestID := uuid.New()
	repo := &requestRepoStub{
		statusID: uuid.New(),
		requests: []CoreIntegrationRequest{{
			ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
			Provider: ProviderSlack, SourceType: "message", SourceExternalID: "Ev-retry",
			Title: "Crash-safe conversion", Priority: "High", Status: StatusPending,
			AcceptanceState: AcceptanceStateIdle,
		}},
	}
	storyService := &idempotentStoryServiceStub{stories: make(map[string]Story)}
	accepter := &flakyProviderAccepterStub{failuresLeft: 1}
	service := New(nil, repo, storyService, map[string]ProviderAccepter{ProviderSlack: accepter})

	_, firstErr := service.Accept(context.Background(), workspaceID, requestID, actorID)
	require.EqualError(t, firstErr, "provider delivery failed")
	require.Equal(t, AcceptanceStateReserved, repo.requests[0].AcceptanceState)
	require.Equal(t, &actorID, repo.requests[0].AcceptanceStartedByUserID)

	_, declineErr := service.Decline(context.Background(), workspaceID, requestID, actorID)
	require.ErrorIs(t, declineErr, ErrRequestNotPending)
	require.Empty(t, repo.markedDeclined)

	// Membership can change after the story commits. A currently authorized
	// teammate must still be able to reconcile the durable reservation without
	// changing its original actor.
	repo.deniedUsers = map[uuid.UUID]struct{}{actorID: {}}
	accepted, err := service.Accept(context.Background(), workspaceID, requestID, retryActorID)
	require.NoError(t, err)
	require.Equal(t, StatusAccepted, accepted.Status)
	require.NotNil(t, accepted.AcceptedStoryID)
	require.Equal(t, 2, storyService.calls, "the retry must reconcile through idempotent story creation")
	require.Len(t, storyService.stories, 1, "only one canonical story may be inserted")
	require.Equal(t, 2, accepter.calls)
	require.Equal(t, accepter.storyIDs[0], accepter.storyIDs[1])
	require.Equal(t, 2, repo.reserveCalls)
	require.Equal(t, &actorID, accepted.AcceptedByUserID)

	again, err := service.Accept(context.Background(), workspaceID, requestID, retryActorID)
	require.NoError(t, err)
	require.Equal(t, accepted.AcceptedStoryID, again.AcceptedStoryID)
	require.Equal(t, 2, storyService.calls)
	require.Equal(t, 2, accepter.calls)
}

func TestAcceptRejectsUnsupportedPriorityBeforeReservation(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	requestID := uuid.New()
	repo := &requestRepoStub{
		statusID: uuid.New(),
		requests: []CoreIntegrationRequest{{
			ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
			Provider: ProviderSlack, SourceType: "message", SourceExternalID: "Ev-priority",
			Title: "Invalid priority", Priority: "Critical", Status: StatusPending,
		}},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{ProviderSlack: providerAccepterStub{}})

	_, err := service.Accept(context.Background(), workspaceID, requestID, actorID)

	require.ErrorIs(t, err, ErrInvalidRequestProperty)
	require.Zero(t, repo.reserveCalls)
	require.Empty(t, repo.createdStories)
}

func TestUpdatePendingRejectsUnsupportedPriority(t *testing.T) {
	repo := &requestRepoStub{}
	service := New(nil, repo, storyServiceStub{repo: repo}, nil)
	priority := "Critical"

	_, err := service.UpdatePending(context.Background(), uuid.New(), uuid.New(), uuid.New(), CoreUpdateRequestInput{Priority: &priority})

	require.ErrorIs(t, err, ErrInvalidRequestProperty)
}

func TestUpsertPendingRejectsUnsupportedPriority(t *testing.T) {
	repo := &requestRepoStub{}
	service := New(nil, repo, storyServiceStub{repo: repo}, nil)

	_, err := service.UpsertPending(context.Background(), CoreUpsertRequestInput{
		WorkspaceID: uuid.New(), TeamID: uuid.New(), Provider: ProviderSlack,
		SourceType: "message", SourceExternalID: "Ev-upsert-priority",
		Title: "Invalid priority", Priority: "Critical",
	})

	require.ErrorIs(t, err, ErrInvalidRequestProperty)
}

func TestListByTeamSupportsPagination(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &requestRepoStub{
		requests: []CoreIntegrationRequest{
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "1", Title: "First", Status: StatusPending},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "2", Title: "Second", Status: StatusPending},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "3", Title: "Third", Status: StatusPending},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "4", Title: "Fourth", Status: StatusPending},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "5", Title: "Fifth", Status: StatusPending},
		},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{})

	requests, err := service.ListByTeam(context.Background(), workspaceID, teamID, actorID, CoreListRequestsFilter{
		Status:   StatusPending,
		Page:     2,
		PageSize: 2,
	})

	require.NoError(t, err)
	require.Len(t, requests, 2)
	require.Equal(t, "Third", requests[0].Title)
	require.Equal(t, "Fourth", requests[1].Title)
}

func TestListByTeamSupportsTriageFilters(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	assigneeID := uuid.New()
	createdAfter := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	repo := &requestRepoStub{
		requests: []CoreIntegrationRequest{
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "1", Title: "Keep", Priority: "Urgent", AssigneeID: &assigneeID, Status: StatusPending, CreatedAt: time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderSlack, SourceType: SourceTypeIssue, SourceExternalID: "2", Title: "Wrong provider", Priority: "Urgent", AssigneeID: &assigneeID, Status: StatusPending, CreatedAt: time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "3", Title: "Wrong priority", Priority: "Low", AssigneeID: &assigneeID, Status: StatusPending, CreatedAt: time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "4", Title: "Too old", Priority: "Urgent", AssigneeID: &assigneeID, Status: StatusPending, CreatedAt: time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)},
		},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{})

	requests, err := service.ListByTeam(context.Background(), workspaceID, teamID, actorID, CoreListRequestsFilter{
		Status:       StatusPending,
		Provider:     ProviderGitHub,
		Priority:     "Urgent",
		AssigneeID:   &assigneeID,
		CreatedAfter: &createdAfter,
	})

	require.NoError(t, err)
	require.Len(t, requests, 1)
	require.Equal(t, "Keep", requests[0].Title)
}

func TestAcceptMapsRequestStoryFieldsToCreatedStory(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	requestID := uuid.New()
	statusID := uuid.New()
	objectiveID := uuid.New()
	keyResultID := uuid.New()
	sprintID := uuid.New()
	estimateValue := int16(5)
	estimatedDurationMinutes := 240
	minimumFocusBlockMinutes := 60
	startDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)
	repo := &requestRepoStub{
		statusID: statusID,
		requests: []CoreIntegrationRequest{
			{
				ID:                       requestID,
				WorkspaceID:              workspaceID,
				TeamID:                   teamID,
				Provider:                 ProviderGitHub,
				SourceType:               SourceTypeIssue,
				SourceExternalID:         "123",
				Title:                    "Import customer escalation",
				StatusID:                 &statusID,
				Priority:                 "Urgent",
				EstimateValue:            &estimateValue,
				EstimatedDurationMinutes: &estimatedDurationMinutes,
				MinimumFocusBlockMinutes: &minimumFocusBlockMinutes,
				ObjectiveID:              &objectiveID,
				KeyResultID:              &keyResultID,
				SprintID:                 &sprintID,
				StartDate:                &startDate,
				EndDate:                  &endDate,
				Status:                   StatusPending,
			},
		},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{
		ProviderGitHub: providerAccepterStub{},
	})

	_, err := service.Accept(context.Background(), workspaceID, requestID, actorID)

	require.NoError(t, err)
	require.Len(t, repo.createdStories, 1)
	createdStory := repo.createdStories[0]
	require.Equal(t, &estimateValue, createdStory.EstimateValue)
	require.Equal(t, &estimatedDurationMinutes, createdStory.EstimatedDurationMinutes)
	require.Equal(t, &minimumFocusBlockMinutes, createdStory.MinimumFocusBlockMinutes)
	require.Equal(t, &objectiveID, createdStory.ObjectiveID)
	require.Equal(t, &keyResultID, createdStory.KeyResultID)
	require.Equal(t, &sprintID, createdStory.SprintID)
	require.Equal(t, &startDate, createdStory.StartDate)
	require.Equal(t, &endDate, createdStory.EndDate)
}

func TestDeclineAllPendingByTeamDeclinesEveryPendingRequest(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	firstID := uuid.New()
	secondID := uuid.New()
	repo := &requestRepoStub{
		requests: []CoreIntegrationRequest{
			{ID: firstID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "1", Title: "First", Status: StatusPending},
			{ID: secondID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderSlack, SourceType: SourceTypeIssue, SourceExternalID: "2", Title: "Second", Status: StatusPending},
			{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderSlack, SourceType: SourceTypeIssue, SourceExternalID: "3", Title: "Handled", Status: StatusAccepted},
		},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{
		ProviderGitHub: providerAccepterStub{},
		ProviderSlack:  providerAccepterStub{},
	})

	result, err := service.DeclineAllPendingByTeam(context.Background(), workspaceID, teamID, actorID)

	require.NoError(t, err)
	require.Equal(t, 2, result.Count)
	require.Equal(t, 2, result.TotalCount)
	require.Equal(t, 2, result.SucceededCount)
	require.Zero(t, result.FailedCount)
	require.False(t, result.Partial)
	require.Len(t, result.Items, 2)
	require.ElementsMatch(t, []uuid.UUID{firstID, secondID}, result.RequestIDs)
	require.ElementsMatch(t, []uuid.UUID{firstID, secondID}, repo.markedDeclined)
}

func TestDeclineAllPendingByTeamContinuesAndReportsPerItemFailures(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	firstID := uuid.New()
	failedID := uuid.New()
	thirdID := uuid.New()
	repo := &requestRepoStub{
		requests: []CoreIntegrationRequest{
			{ID: firstID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderGitHub, SourceType: SourceTypeIssue, SourceExternalID: "1", Title: "First", Status: StatusPending},
			{ID: failedID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderSlack, SourceType: SourceTypeIssue, SourceExternalID: "2", Title: "Reserved", Status: StatusPending, AcceptanceState: AcceptanceStateReserved},
			{ID: thirdID, WorkspaceID: workspaceID, TeamID: teamID, Provider: ProviderSlack, SourceType: SourceTypeIssue, SourceExternalID: "3", Title: "Third", Status: StatusPending},
		},
	}
	service := New(nil, repo, storyServiceStub{repo: repo}, map[string]ProviderAccepter{
		ProviderGitHub: providerAccepterStub{},
		ProviderSlack:  providerAccepterStub{},
	})

	result, err := service.DeclineAllPendingByTeam(context.Background(), workspaceID, teamID, actorID)

	require.NoError(t, err)
	require.Equal(t, 3, result.TotalCount)
	require.Equal(t, 2, result.SucceededCount)
	require.Equal(t, 1, result.FailedCount)
	require.True(t, result.Partial)
	require.Equal(t, []uuid.UUID{firstID, thirdID}, result.RequestIDs)
	require.Len(t, result.Items, 3)
	require.Equal(t, firstID, result.Items[0].RequestID)
	require.True(t, result.Items[0].Success)
	require.Equal(t, failedID, result.Items[1].RequestID)
	require.False(t, result.Items[1].Success)
	require.Equal(t, ErrRequestNotPending.Error(), result.Items[1].Error)
	require.Equal(t, thirdID, result.Items[2].RequestID)
	require.True(t, result.Items[2].Success)
}
