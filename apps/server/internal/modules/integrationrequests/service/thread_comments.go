package integrationrequests

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) BindProviderThread(ctx context.Context, input CoreBindProviderThreadInput) (CoreProviderThread, error) {
	if input.WorkspaceID == uuid.Nil || input.IntegrationRequestID == uuid.Nil {
		return CoreProviderThread{}, errors.New("workspace and integration request are required")
	}
	if strings.TrimSpace(input.Provider) == "" || strings.TrimSpace(input.ExternalWorkspaceID) == "" || strings.TrimSpace(input.ExternalChannelID) == "" || strings.TrimSpace(input.ExternalThreadID) == "" {
		return CoreProviderThread{}, errors.New("provider workspace, channel, and thread are required")
	}
	return s.repo.BindProviderThread(ctx, input)
}

func (s *Service) HasAuthorizedProviderThread(ctx context.Context, input CoreProviderThreadMatchInput) (bool, error) {
	return s.repo.HasAuthorizedProviderThread(ctx, input)
}

func (s *Service) HasCurrentProviderThread(ctx context.Context, input CoreProviderThreadLookupInput) (bool, error) {
	return s.repo.HasCurrentProviderThread(ctx, input)
}

// FindProviderThread is reserved for a provider accepter that has already
// received an actor-authorized request from this service.
func (s *Service) FindProviderThread(ctx context.Context, workspaceID, requestID uuid.UUID, provider string) (CoreProviderThread, error) {
	return s.repo.FindProviderThread(ctx, workspaceID, requestID, provider)
}

func (s *Service) GetThreadActivityForRequest(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (CoreThreadActivity, error) {
	return s.repo.GetThreadActivityForRequest(ctx, workspaceID, requestID, userID)
}

func (s *Service) ListProviderThreadsForStory(ctx context.Context, workspaceID, storyID, userID uuid.UUID) ([]CoreProviderThread, error) {
	return s.repo.ListProviderThreadsForStory(ctx, workspaceID, storyID, userID)
}

func (s *Service) CreateComment(ctx context.Context, input CoreCreateCommentInput) (CoreIntegrationRequestComment, error) {
	input.Body = strings.TrimSpace(input.Body)
	if input.WorkspaceID == uuid.Nil || input.RequestID == uuid.Nil || input.AuthorID == uuid.Nil {
		return CoreIntegrationRequestComment{}, errors.New("workspace, integration request, and author are required")
	}
	if input.Body == "" {
		return CoreIntegrationRequestComment{}, fmt.Errorf("%w: comment body is required", ErrInvalidRequestProperty)
	}
	if input.ClientIdempotencyKey == uuid.Nil {
		return CoreIntegrationRequestComment{}, fmt.Errorf("%w: comment idempotency key is required", ErrInvalidRequestProperty)
	}
	request, err := s.repo.GetForUser(ctx, input.WorkspaceID, input.RequestID, input.AuthorID)
	if err != nil {
		return CoreIntegrationRequestComment{}, err
	}
	thread, err := s.repo.FindProviderThread(ctx, input.WorkspaceID, input.RequestID, request.Provider)
	if err != nil {
		return CoreIntegrationRequestComment{}, err
	}
	commenter := s.providerCommenters[thread.Provider]
	if commenter == nil {
		return CoreIntegrationRequestComment{}, fmt.Errorf("%w: %s", ErrUnsupportedProvider, thread.Provider)
	}
	prepared, err := commenter.PrepareIntegrationRequestComment(ctx, request, thread, input)
	if err != nil {
		return CoreIntegrationRequestComment{}, err
	}
	thread, comment, err := s.repo.CreateOutboundComment(ctx, input, prepared)
	if err != nil {
		return CoreIntegrationRequestComment{}, err
	}
	if comment.DeliveryStatus != nil && (*comment.DeliveryStatus == "sent" || *comment.DeliveryStatus == "not-sent" || *comment.DeliveryStatus == "failed") {
		return comment, nil
	}
	if err := commenter.DeliverIntegrationRequestComment(ctx, request, thread, comment, prepared); err != nil && s.log != nil {
		s.log.Warn(ctx, "integration request comment queued after immediate provider delivery failed", "error", err, "request_id", request.ID, "comment_id", comment.ID, "provider", thread.Provider)
	}
	refreshed, err := s.repo.GetCommentForUser(ctx, input.WorkspaceID, comment.ID, input.AuthorID)
	if err == nil {
		return refreshed, nil
	}
	if s.log != nil {
		s.log.Warn(ctx, "failed refreshing integration request comment delivery state", "error", err, "request_id", request.ID, "comment_id", comment.ID)
	}
	return comment, nil
}
