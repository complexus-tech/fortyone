package integrationrequests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type requestRepoStub struct {
	requests       []CoreIntegrationRequest
	markedAccepted []uuid.UUID
	markedDeclined []uuid.UUID
	statusID       uuid.UUID
	createdStories []stories.CoreNewStory
}

func (r *requestRepoStub) UpsertPending(ctx context.Context, input CoreUpsertRequestInput) (CoreIntegrationRequest, error) {
	return CoreIntegrationRequest{}, nil
}

func (r *requestRepoStub) ListByTeam(ctx context.Context, workspaceID, teamID uuid.UUID, filter CoreListRequestsFilter) ([]CoreIntegrationRequest, error) {
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

func (r *requestRepoStub) CountByTeam(ctx context.Context, workspaceID, teamID uuid.UUID, filter CoreListRequestsFilter) (int, error) {
	filter.Page = 0
	filter.PageSize = 0
	requests, err := r.ListByTeam(ctx, workspaceID, teamID, filter)
	return len(requests), err
}

func (r *requestRepoStub) Get(ctx context.Context, workspaceID, requestID uuid.UUID) (CoreIntegrationRequest, error) {
	for _, request := range r.requests {
		if request.WorkspaceID == workspaceID && request.ID == requestID {
			return request, nil
		}
	}
	return CoreIntegrationRequest{}, sql.ErrNoRows
}

func (r *requestRepoStub) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	return &r.statusID, nil
}

func (r *requestRepoStub) UpdatePending(ctx context.Context, workspaceID, requestID uuid.UUID, input CoreUpdateRequestInput) (CoreIntegrationRequest, error) {
	return CoreIntegrationRequest{}, nil
}

func (r *requestRepoStub) MarkAccepted(ctx context.Context, workspaceID, requestID, storyID, acceptedByUserID uuid.UUID) (CoreIntegrationRequest, error) {
	r.markedAccepted = append(r.markedAccepted, requestID)
	request, err := r.Get(ctx, workspaceID, requestID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	request.Status = StatusAccepted
	request.AcceptedStoryID = &storyID
	return request, nil
}

func (r *requestRepoStub) MarkDeclined(ctx context.Context, workspaceID, requestID, declinedByUserID uuid.UUID) (CoreIntegrationRequest, error) {
	r.markedDeclined = append(r.markedDeclined, requestID)
	request, err := r.Get(ctx, workspaceID, requestID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	request.Status = StatusDeclined
	return request, nil
}

type storyServiceStub struct {
	repo *requestRepoStub
}

func (s storyServiceStub) CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	s.repo.createdStories = append(s.repo.createdStories, ns)
	return stories.CoreSingleStory{ID: uuid.New()}, nil
}

type providerAccepterStub struct{}

func (providerAccepterStub) AcceptIntegrationRequest(ctx context.Context, request CoreIntegrationRequest, story stories.CoreSingleStory) error {
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
}

func TestListByTeamSupportsPagination(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
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

	requests, err := service.ListByTeam(context.Background(), workspaceID, teamID, CoreListRequestsFilter{
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

	requests, err := service.ListByTeam(context.Background(), workspaceID, teamID, CoreListRequestsFilter{
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
	startDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)
	repo := &requestRepoStub{
		statusID: statusID,
		requests: []CoreIntegrationRequest{
			{
				ID:               requestID,
				WorkspaceID:      workspaceID,
				TeamID:           teamID,
				Provider:         ProviderGitHub,
				SourceType:       SourceTypeIssue,
				SourceExternalID: "123",
				Title:            "Import customer escalation",
				StatusID:         &statusID,
				Priority:         "Urgent",
				EstimateValue:    &estimateValue,
				ObjectiveID:      &objectiveID,
				KeyResultID:      &keyResultID,
				SprintID:         &sprintID,
				StartDate:        &startDate,
				EndDate:          &endDate,
				Status:           StatusPending,
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
