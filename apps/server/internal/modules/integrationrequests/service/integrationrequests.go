package integrationrequests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrUnsupportedProvider    = errors.New("unsupported integration request provider")
	ErrRequestNotPending      = errors.New("integration request is not pending")
	ErrInvalidRequestProperty = errors.New("invalid integration request property")
	ErrProviderThreadNotFound = errors.New("integration request provider thread not found")
	ErrIdempotencyConflict    = errors.New("comment idempotency key was already used for different content")
)

var supportedRequestPriorities = map[string]string{
	"No Priority": "No Priority",
	"Low":         "Low",
	"Medium":      "Medium",
	"High":        "High",
	"Urgent":      "Urgent",
}

type Repository interface {
	UpsertPending(ctx context.Context, input CoreUpsertRequestInput) (CoreIntegrationRequest, error)
	AuthorizeTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID) error
	ListByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter CoreListRequestsFilter) ([]CoreIntegrationRequest, error)
	CountByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter CoreListRequestsFilter) (int, error)
	GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (CoreIntegrationRequest, error)
	FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error)
	UpdatePending(ctx context.Context, workspaceID, requestID, userID uuid.UUID, input CoreUpdateRequestInput) (CoreIntegrationRequest, error)
	ReserveAcceptance(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (CoreIntegrationRequest, error)
	MarkAccepted(ctx context.Context, workspaceID, requestID, storyID, acceptedByUserID uuid.UUID) (CoreIntegrationRequest, error)
	MarkDeclined(ctx context.Context, workspaceID, requestID, declinedByUserID uuid.UUID) (CoreIntegrationRequest, error)
	BindProviderThread(ctx context.Context, input CoreBindProviderThreadInput) (CoreProviderThread, error)
	HasAuthorizedProviderThread(ctx context.Context, input CoreProviderThreadMatchInput) (bool, error)
	HasCurrentProviderThread(ctx context.Context, input CoreProviderThreadLookupInput) (bool, error)
	FindProviderThread(ctx context.Context, workspaceID, requestID uuid.UUID, provider string) (CoreProviderThread, error)
	GetThreadActivityForRequest(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (CoreThreadActivity, error)
	ListProviderThreadsForStory(ctx context.Context, workspaceID, storyID, userID uuid.UUID) ([]CoreProviderThread, error)
	CreateOutboundComment(ctx context.Context, input CoreCreateCommentInput, prepared CorePreparedProviderComment) (CoreProviderThread, CoreIntegrationRequestComment, error)
	GetCommentForUser(ctx context.Context, workspaceID, commentID, userID uuid.UUID) (CoreIntegrationRequestComment, error)
}

type Service struct {
	log                *logger.Logger
	repo               Repository
	stories            StoryService
	providerAccepters  map[string]ProviderAccepter
	providerCommenters map[string]ProviderCommenter
}

type Option func(*Service)

func WithProviderCommenter(provider string, commenter ProviderCommenter) Option {
	return func(service *Service) {
		provider = strings.TrimSpace(provider)
		if provider != "" && commenter != nil {
			service.providerCommenters[provider] = commenter
		}
	}
}

func New(log *logger.Logger, repo Repository, stories StoryService, providerAccepters map[string]ProviderAccepter, options ...Option) *Service {
	service := &Service{
		log:                log,
		repo:               repo,
		stories:            stories,
		providerAccepters:  providerAccepters,
		providerCommenters: make(map[string]ProviderCommenter),
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *Service) UpsertPending(ctx context.Context, input CoreUpsertRequestInput) (CoreIntegrationRequest, error) {
	if err := validateUpsertInput(input); err != nil {
		return CoreIntegrationRequest{}, err
	}
	priority, err := normalizeRequestPriority(input.Priority)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	input.Priority = priority
	return s.repo.UpsertPending(ctx, input)
}

func (s *Service) ListByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter CoreListRequestsFilter) ([]CoreIntegrationRequest, error) {
	if err := s.repo.AuthorizeTeam(ctx, workspaceID, teamID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByTeam(ctx, workspaceID, teamID, userID, filter)
}

func (s *Service) CountByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter CoreListRequestsFilter) (int, error) {
	if err := s.repo.AuthorizeTeam(ctx, workspaceID, teamID, userID); err != nil {
		return 0, err
	}
	return s.repo.CountByTeam(ctx, workspaceID, teamID, userID, filter)
}

// Get resolves an integration request for the authenticated actor stored in
// the context. Provider-specific HTTP services use this compatibility method;
// callers that already have the actor ID should use GetForUser directly.
func (s *Service) Get(ctx context.Context, workspaceID, requestID uuid.UUID) (CoreIntegrationRequest, error) {
	userID, err := platformauth.GetUserID(ctx)
	if err != nil {
		return CoreIntegrationRequest{}, fmt.Errorf("resolve integration request actor: %w", err)
	}
	return s.GetForUser(ctx, workspaceID, requestID, userID)
}

func (s *Service) GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (CoreIntegrationRequest, error) {
	return s.repo.GetForUser(ctx, workspaceID, requestID, userID)
}

func (s *Service) UpdatePending(ctx context.Context, workspaceID, requestID, userID uuid.UUID, input CoreUpdateRequestInput) (CoreIntegrationRequest, error) {
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return CoreIntegrationRequest{}, fmt.Errorf("%w: title is required", ErrInvalidRequestProperty)
	}
	if input.Priority != nil {
		priority, err := normalizeRequestPriority(*input.Priority)
		if err != nil {
			return CoreIntegrationRequest{}, err
		}
		input.Priority = &priority
	}
	if input.EstimatedDurationMinutes.Set && input.EstimatedDurationMinutes.Value != nil && *input.EstimatedDurationMinutes.Value <= 0 {
		return CoreIntegrationRequest{}, fmt.Errorf("%w: %w", ErrInvalidRequestProperty, stories.ErrInvalidEstimatedDuration)
	}
	if input.EstimatedDurationMinutes.Set && input.EstimatedDurationMinutes.Value != nil && *input.EstimatedDurationMinutes.Value > stories.MaximumEstimatedDurationMinutes {
		return CoreIntegrationRequest{}, fmt.Errorf("%w: %w", ErrInvalidRequestProperty, stories.ErrEstimatedDurationTooLarge)
	}
	if input.MinimumFocusBlockMinutes.Set && input.MinimumFocusBlockMinutes.Value != nil && *input.MinimumFocusBlockMinutes.Value <= 0 {
		return CoreIntegrationRequest{}, fmt.Errorf("%w: %w", ErrInvalidRequestProperty, stories.ErrInvalidMinimumFocusBlock)
	}
	if input.MinimumFocusBlockMinutes.Set && input.MinimumFocusBlockMinutes.Value != nil && *input.MinimumFocusBlockMinutes.Value > stories.MaximumEstimatedDurationMinutes {
		return CoreIntegrationRequest{}, fmt.Errorf("%w: %w", ErrInvalidRequestProperty, stories.ErrMinimumFocusBlockTooLarge)
	}
	return s.repo.UpdatePending(ctx, workspaceID, requestID, userID, input)
}

func (s *Service) Accept(ctx context.Context, workspaceID, requestID, actorID uuid.UUID) (CoreIntegrationRequest, error) {
	request, err := s.repo.GetForUser(ctx, workspaceID, requestID, actorID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	if request.Status == StatusAccepted && request.AcceptedStoryID != nil {
		return request, nil
	}
	if request.Status != StatusPending {
		return CoreIntegrationRequest{}, ErrRequestNotPending
	}
	if _, err := normalizeRequestPriority(request.Priority); err != nil {
		return CoreIntegrationRequest{}, err
	}
	accepter := s.providerAccepters[request.Provider]
	if accepter == nil {
		return CoreIntegrationRequest{}, fmt.Errorf("%w: %s", ErrUnsupportedProvider, request.Provider)
	}

	// Reserve conversion before creating the story. The repository locks and
	// revalidates the pending row, then durably fences edits and declines. A
	// retry resumes the same reservation and relies on CreationKey for exactly
	// one story even after a process crash.
	request, err = s.repo.ReserveAcceptance(ctx, workspaceID, requestID, actorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			current, getErr := s.repo.GetForUser(ctx, workspaceID, requestID, actorID)
			if getErr == nil && current.Status == StatusAccepted && current.AcceptedStoryID != nil {
				return current, nil
			}
			if getErr == nil {
				return CoreIntegrationRequest{}, ErrRequestNotPending
			}
		}
		return CoreIntegrationRequest{}, err
	}
	conversionActorID := actorID
	if request.AcceptanceStartedByUserID != nil {
		conversionActorID = *request.AcceptanceStartedByUserID
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
	priority, err := normalizeRequestPriority(request.Priority)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	creationKey := fmt.Sprintf("integration-request:%s:%s", workspaceID, request.ID)

	story, err := s.stories.CreateExternalUserAction(ctx, conversionActorID, stories.CoreNewStory{
		Title:                    request.Title,
		Description:              request.Description,
		Status:                   statusID,
		Reporter:                 &conversionActorID,
		Assignee:                 request.AssigneeID,
		Team:                     request.TeamID,
		Priority:                 priority,
		EstimateValue:            request.EstimateValue,
		EstimatedDurationMinutes: request.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: request.MinimumFocusBlockMinutes,
		Objective:                request.ObjectiveID,
		KeyResult:                request.KeyResultID,
		Sprint:                   request.SprintID,
		StartDate:                request.StartDate,
		EndDate:                  request.EndDate,
		LabelIDs:                 request.LabelIDs,
		CreationKey:              &creationKey,
	}, workspaceID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}

	if err := accepter.AcceptIntegrationRequest(ctx, request, story); err != nil {
		return CoreIntegrationRequest{}, err
	}

	accepted, err := s.repo.MarkAccepted(ctx, workspaceID, requestID, story.ID, conversionActorID)
	if err == nil {
		return accepted, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return CoreIntegrationRequest{}, err
	}
	current, getErr := s.repo.GetForUser(ctx, workspaceID, requestID, actorID)
	if getErr == nil && current.Status == StatusAccepted && current.AcceptedStoryID != nil && *current.AcceptedStoryID == story.ID {
		return current, nil
	}
	return CoreIntegrationRequest{}, err
}

func (s *Service) AcceptAllPendingByTeam(ctx context.Context, workspaceID, teamID, actorID uuid.UUID) (CoreBulkRequestResult, error) {
	requests, err := s.ListByTeam(ctx, workspaceID, teamID, actorID, CoreListRequestsFilter{Status: StatusPending})
	if err != nil {
		return CoreBulkRequestResult{}, err
	}

	result := CoreBulkRequestResult{
		TotalCount: len(requests),
		RequestIDs: make([]uuid.UUID, 0, len(requests)),
		Items:      make([]CoreBulkRequestItemResult, 0, len(requests)),
	}
	for _, request := range requests {
		accepted, err := s.Accept(ctx, workspaceID, request.ID, actorID)
		if err != nil {
			result.FailedCount++
			result.Items = append(result.Items, CoreBulkRequestItemResult{
				RequestID: request.ID,
				Status:    "failed",
				Error:     err.Error(),
			})
			continue
		}
		result.SucceededCount++
		result.RequestIDs = append(result.RequestIDs, request.ID)
		result.Items = append(result.Items, CoreBulkRequestItemResult{
			RequestID:       request.ID,
			Success:         true,
			Status:          StatusAccepted,
			AcceptedStoryID: accepted.AcceptedStoryID,
		})
	}
	result.Count = result.SucceededCount
	result.Partial = result.SucceededCount > 0 && result.FailedCount > 0
	return result, nil
}

func (s *Service) Decline(ctx context.Context, workspaceID, requestID, actorID uuid.UUID) (CoreIntegrationRequest, error) {
	request, err := s.repo.GetForUser(ctx, workspaceID, requestID, actorID)
	if err != nil {
		return CoreIntegrationRequest{}, err
	}
	if request.Status != StatusPending || request.AcceptanceState == AcceptanceStateReserved {
		return CoreIntegrationRequest{}, ErrRequestNotPending
	}
	return s.repo.MarkDeclined(ctx, workspaceID, requestID, actorID)
}

func (s *Service) DeclineAllPendingByTeam(ctx context.Context, workspaceID, teamID, actorID uuid.UUID) (CoreBulkRequestResult, error) {
	requests, err := s.ListByTeam(ctx, workspaceID, teamID, actorID, CoreListRequestsFilter{Status: StatusPending})
	if err != nil {
		return CoreBulkRequestResult{}, err
	}

	result := CoreBulkRequestResult{
		TotalCount: len(requests),
		RequestIDs: make([]uuid.UUID, 0, len(requests)),
		Items:      make([]CoreBulkRequestItemResult, 0, len(requests)),
	}
	for _, request := range requests {
		if _, err := s.Decline(ctx, workspaceID, request.ID, actorID); err != nil {
			result.FailedCount++
			result.Items = append(result.Items, CoreBulkRequestItemResult{
				RequestID: request.ID,
				Status:    "failed",
				Error:     err.Error(),
			})
			continue
		}
		result.SucceededCount++
		result.RequestIDs = append(result.RequestIDs, request.ID)
		result.Items = append(result.Items, CoreBulkRequestItemResult{
			RequestID: request.ID,
			Success:   true,
			Status:    StatusDeclined,
		})
	}
	result.Count = result.SucceededCount
	result.Partial = result.SucceededCount > 0 && result.FailedCount > 0
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
	if err := stories.ValidateStoryTimeContract(input.EstimatedDurationMinutes, input.MinimumFocusBlockMinutes); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRequestProperty, err)
	}
	return nil
}

func normalizeRequestPriority(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "No Priority", nil
	}
	if priority, ok := supportedRequestPriorities[trimmed]; ok {
		return priority, nil
	}
	return "", fmt.Errorf("%w: unsupported priority %q", ErrInvalidRequestProperty, trimmed)
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
