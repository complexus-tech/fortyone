package integrationrequests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported integration request provider")
	ErrRequestNotPending   = errors.New("integration request is not pending")
)

type Repository interface {
	UpsertPending(ctx context.Context, input CoreUpsertRequestInput) (CoreIntegrationRequest, error)
	ListByTeam(ctx context.Context, workspaceID, teamID uuid.UUID, filter CoreListRequestsFilter) ([]CoreIntegrationRequest, error)
	Get(ctx context.Context, workspaceID, requestID uuid.UUID) (CoreIntegrationRequest, error)
	FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error)
	UpdatePending(ctx context.Context, workspaceID, requestID uuid.UUID, input CoreUpdateRequestInput) (CoreIntegrationRequest, error)
	MarkAccepted(ctx context.Context, workspaceID, requestID, storyID, acceptedByUserID uuid.UUID) (CoreIntegrationRequest, error)
	MarkDeclined(ctx context.Context, workspaceID, requestID, declinedByUserID uuid.UUID) (CoreIntegrationRequest, error)
}

type Service struct {
	log               *logger.Logger
	repo              Repository
	stories           StoryService
	providerAccepters map[string]ProviderAccepter
}

func New(log *logger.Logger, repo Repository, stories StoryService, providerAccepters map[string]ProviderAccepter) *Service {
	return &Service{
		log:               log,
		repo:              repo,
		stories:           stories,
		providerAccepters: providerAccepters,
	}
}

func (s *Service) UpsertPending(ctx context.Context, input CoreUpsertRequestInput) (CoreIntegrationRequest, error) {
	if err := validateUpsertInput(input); err != nil {
		return CoreIntegrationRequest{}, err
	}
	return s.repo.UpsertPending(ctx, input)
}

func (s *Service) ListByTeam(ctx context.Context, workspaceID, teamID uuid.UUID, filter CoreListRequestsFilter) ([]CoreIntegrationRequest, error) {
	return s.repo.ListByTeam(ctx, workspaceID, teamID, filter)
}

func (s *Service) Get(ctx context.Context, workspaceID, requestID uuid.UUID) (CoreIntegrationRequest, error) {
	return s.repo.Get(ctx, workspaceID, requestID)
}

func (s *Service) UpdatePending(ctx context.Context, workspaceID, requestID uuid.UUID, input CoreUpdateRequestInput) (CoreIntegrationRequest, error) {
	return s.repo.UpdatePending(ctx, workspaceID, requestID, input)
}

func (s *Service) Accept(ctx context.Context, workspaceID, requestID, actorID uuid.UUID) (CoreIntegrationRequest, error) {
	request, err := s.repo.Get(ctx, workspaceID, requestID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	if request.Status != StatusPending {
		return CoreIntegrationRequest{}, ErrRequestNotPending
	}
	accepter := s.providerAccepters[request.Provider]
	if accepter == nil {
		return CoreIntegrationRequest{}, fmt.Errorf("%w: %s", ErrUnsupportedProvider, request.Provider)
	}

	statusID := request.StatusID
	if statusID == nil {
		var err error
		statusID, err = s.repo.FindFirstStatusByCategory(ctx, request.TeamID, "unstarted")
		if err != nil {
			return CoreIntegrationRequest{}, err
		}
		if statusID == nil {
			return CoreIntegrationRequest{}, errors.New("team has no unstarted status configured")
		}
	}
	priority := strings.TrimSpace(request.Priority)
	if priority == "" {
		priority = "No Priority"
	}

	story, err := s.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title:       request.Title,
		Description: request.Description,
		Status:      statusID,
		Reporter:    &actorID,
		Assignee:    request.AssigneeID,
		Team:        request.TeamID,
		Priority:    priority,
	}, workspaceID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}

	if err := accepter.AcceptIntegrationRequest(ctx, request, story); err != nil {
		return CoreIntegrationRequest{}, err
	}

	return s.repo.MarkAccepted(ctx, workspaceID, requestID, story.ID, actorID)
}

func (s *Service) AcceptAllPendingByTeam(ctx context.Context, workspaceID, teamID, actorID uuid.UUID) (CoreBulkRequestResult, error) {
	requests, err := s.repo.ListByTeam(ctx, workspaceID, teamID, CoreListRequestsFilter{Status: StatusPending})
	if err != nil {
		return CoreBulkRequestResult{}, err
	}

	result := CoreBulkRequestResult{
		RequestIDs: make([]uuid.UUID, 0, len(requests)),
	}
	for _, request := range requests {
		if _, err := s.Accept(ctx, workspaceID, request.ID, actorID); err != nil {
			return CoreBulkRequestResult{}, err
		}
		result.RequestIDs = append(result.RequestIDs, request.ID)
	}
	result.Count = len(result.RequestIDs)
	return result, nil
}

func (s *Service) Decline(ctx context.Context, workspaceID, requestID, actorID uuid.UUID) (CoreIntegrationRequest, error) {
	request, err := s.repo.Get(ctx, workspaceID, requestID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	if request.Status != StatusPending {
		return CoreIntegrationRequest{}, ErrRequestNotPending
	}
	return s.repo.MarkDeclined(ctx, workspaceID, requestID, actorID)
}

func (s *Service) DeclineAllPendingByTeam(ctx context.Context, workspaceID, teamID, actorID uuid.UUID) (CoreBulkRequestResult, error) {
	requests, err := s.repo.ListByTeam(ctx, workspaceID, teamID, CoreListRequestsFilter{Status: StatusPending})
	if err != nil {
		return CoreBulkRequestResult{}, err
	}

	result := CoreBulkRequestResult{
		RequestIDs: make([]uuid.UUID, 0, len(requests)),
	}
	for _, request := range requests {
		if _, err := s.Decline(ctx, workspaceID, request.ID, actorID); err != nil {
			return CoreBulkRequestResult{}, err
		}
		result.RequestIDs = append(result.RequestIDs, request.ID)
	}
	result.Count = len(result.RequestIDs)
	return result, nil
}

func validateUpsertInput(input CoreUpsertRequestInput) error {
	if input.WorkspaceID == uuid.Nil {
		return errors.New("workspace id is required")
	}
	if input.TeamID == uuid.Nil {
		return errors.New("team id is required")
	}
	if strings.TrimSpace(input.Provider) == "" {
		return errors.New("provider is required")
	}
	if strings.TrimSpace(input.SourceType) == "" {
		return errors.New("source type is required")
	}
	if strings.TrimSpace(input.SourceExternalID) == "" {
		return errors.New("source external id is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
