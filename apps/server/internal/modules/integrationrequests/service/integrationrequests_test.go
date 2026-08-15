package integrationrequests

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type requestRepoStub struct {
	requests        []CoreIntegrationRequest
	markedAccepted  []uuid.UUID
	markedDeclined  []uuid.UUID
	statusID        uuid.UUID
	createdStories  []stories.CoreNewStory
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
		return sql.ErrNoRows
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
	return CoreIntegrationRequest{}, sql.ErrNoRows
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
		return CoreIntegrationRequest{}, sql.ErrNoRows
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
		return CoreIntegrationRequest{}, sql.ErrNoRows
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
		return CoreIntegrationRequest{}, sql.ErrNoRows
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
	stories map[string]stories.CoreSingleStory
}

func (s *idempotentStoryServiceStub) CreateExternalUserAction(_ context.Context, _ uuid.UUID, input stories.CoreNewStory, _ uuid.UUID) (stories.CoreSingleStory, error) {
	s.calls++
	if input.CreationKey == nil {
		return stories.CoreSingleStory{}, errors.New("creation key is required")
	}
	if existing, ok := s.stories[*input.CreationKey]; ok {
		existing.CreatedNow = false
		return existing, nil
	}
	created := stories.CoreSingleStory{ID: uuid.New(), CreatedNow: true}
	s.stories[*input.CreationKey] = created
	return created, nil
}

func (s storyServiceStub) CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	s.repo.createdStories = append(s.repo.createdStories, ns)
	return stories.CoreSingleStory{ID: uuid.New()}, nil
}

type providerAccepterStub struct{}

func (providerAccepterStub) AcceptIntegrationRequest(ctx context.Context, request CoreIntegrationRequest, story stories.CoreSingleStory) error {
	return nil
}

type flakyProviderAccepterStub struct {
	calls        int
	failuresLeft int
	storyIDs     []uuid.UUID
}

func (s *flakyProviderAccepterStub) AcceptIntegrationRequest(_ context.Context, _ CoreIntegrationRequest, story stories.CoreSingleStory) error {
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
	require.ElementsMatch(t, []uuid.UUID{firstID, secondID}, result.RequestIDs)
	require.ElementsMatch(t, []uuid.UUID{firstID, secondID}, repo.markedAccepted)
	require.Len(t, repo.createdStories, 2)
	require.Equal(t, "High", repo.createdStories[0].Priority)
	require.Equal(t, "No Priority", repo.createdStories[1].Priority)
	for index, created := range repo.createdStories {
		require.NotNil(t, created.CreationKey, "story %d has no idempotency key", index)
		require.Contains(t, *created.CreationKey, workspaceID.String())
	}
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
	require.NotNil(t, repo.createdStories[0].CreationKey)
	require.Equal(t, "integration-request:"+workspaceID.String()+":"+requestID.String(), *repo.createdStories[0].CreationKey)
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
	storyService := &idempotentStoryServiceStub{stories: make(map[string]stories.CoreSingleStory)}
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
	require.Equal(t, &objectiveID, createdStory.Objective)
	require.Equal(t, &keyResultID, createdStory.KeyResult)
	require.Equal(t, &sprintID, createdStory.Sprint)
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
	require.ElementsMatch(t, []uuid.UUID{firstID, secondID}, result.RequestIDs)
	require.ElementsMatch(t, []uuid.UUID{firstID, secondID}, repo.markedDeclined)
}

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
			name: "context get",
			run: func(service *Service) error {
				ctx := platformauth.SetUserID(context.Background(), actorID)
				_, err := service.Get(ctx, workspaceID, requestID)
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

			require.ErrorIs(t, err, sql.ErrNoRows)
			require.Empty(t, repo.createdStories)
			require.Empty(t, repo.markedAccepted)
			require.Empty(t, repo.markedDeclined)
		})
	}
}

func TestGetRequiresAuthenticatedActorContext(t *testing.T) {
	service := New(nil, &requestRepoStub{}, storyServiceStub{}, nil)

	_, err := service.Get(context.Background(), uuid.New(), uuid.New())

	require.Error(t, err)
	require.Contains(t, err.Error(), "resolve integration request actor")
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
