package stories

import (
	"context"
	"errors"
	"fmt"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

// storySupportReadRepository contains the small, actor-scoped reads used by
// story subresources and validation. The repository receives explicit
// authority; it never reaches into request context to infer identity.
type storySupportReadRepository interface {
	GetTeamEstimateScheme(context.Context, StoryReadScope, uuid.UUID) (string, error)
	FindFirstStatusByCategory(context.Context, StoryReadScope, uuid.UUID, string) (*uuid.UUID, error)
	ResolveKeyResult(context.Context, StoryReadScope, uuid.UUID) (storydomain.MutationKeyResultReference, error)
	ListVisibleStoryLinks(context.Context, StoryReadScope, uuid.UUID) ([]storydomain.StoryLink, error)
	ListVisibleStoryActivities(context.Context, StoryReadScope, uuid.UUID, int, int) ([]storydomain.ActivityWithUser, bool, error)
	GetStatusCategory(context.Context, StoryReadScope, uuid.UUID) (string, error)
}

func (s *Service) storySupportReads() (storySupportReadRepository, error) {
	repository, ok := s.repo.(storySupportReadRepository)
	if !ok {
		return nil, errors.New("story repository does not support scoped support reads")
	}
	return repository, nil
}

func (s *Service) resolveStoryKeyResult(
	ctx context.Context,
	workspaceID, keyResultID uuid.UUID,
) (CoreKeyResultReference, error) {
	repository, err := s.storySupportReads()
	if err != nil {
		legacy, ok := s.repo.(legacyKeyResultRepository)
		if !ok {
			return CoreKeyResultReference{}, err
		}
		return legacy.ResolveKeyResult(ctx, keyResultID, workspaceID)
	}
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return CoreKeyResultReference{}, err
	}
	keyResult, err := repository.ResolveKeyResult(ctx, scope, keyResultID)
	if err != nil {
		return CoreKeyResultReference{}, err
	}
	return CoreKeyResultReference{ObjectiveID: keyResult.ObjectiveID, Name: keyResult.Name}, nil
}

func (s *Service) getTeamEstimateScheme(
	ctx context.Context,
	workspaceID, teamID uuid.UUID,
) (string, error) {
	repository, err := s.storySupportReads()
	if err != nil {
		legacy, ok := s.repo.(legacyEstimateSchemeRepository)
		if !ok {
			return "", err
		}
		return legacy.GetTeamEstimateScheme(ctx, teamID, workspaceID)
	}
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	scheme, err := repository.GetTeamEstimateScheme(ctx, scope, teamID)
	if err != nil {
		return "", err
	}
	if err := ValidateEstimateScheme(scheme); err != nil {
		return DefaultEstimateScheme, nil
	}
	return scheme, nil
}

func (s *Service) getStoryStatusCategory(
	ctx context.Context,
	workspaceID, statusID uuid.UUID,
) (string, error) {
	repository, err := s.storySupportReads()
	if err != nil {
		legacy, ok := s.repo.(legacyStatusCategoryRepository)
		if !ok {
			return "", err
		}
		return legacy.GetStatusCategory(ctx, statusID.String())
	}
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	category, err := repository.GetStatusCategory(ctx, scope, statusID)
	if err != nil {
		return "", fmt.Errorf("get story status category: %w", err)
	}
	return category, nil
}

func coreActivityWithUser(activity storydomain.ActivityWithUser) CoreActivityWithUser {
	return CoreActivityWithUser{
		ID: activity.ID, StoryID: activity.StoryID, UserID: activity.UserID,
		Type: activity.Type, Field: activity.Field, CurrentValue: activity.CurrentValue,
		OldValue: activity.OldValue, NewValue: activity.NewValue, Reason: activity.Reason,
		CreatedAt: activity.CreatedAt, WorkspaceID: activity.WorkspaceID,
		User: UserDetails{
			ID: activity.User.ID, Username: activity.User.Username, FullName: activity.User.FullName,
			AvatarURL: activity.User.AvatarURL, IsActive: activity.User.IsActive, IsSystem: activity.User.IsSystem,
		},
	}
}
