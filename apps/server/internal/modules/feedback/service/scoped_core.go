package feedback

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Service) coreScope(ctx context.Context, workspaceID, fallbackActorID uuid.UUID) (CoreAccessScope, bool, error) {
	if s.scopedCoreRepo == nil {
		return CoreAccessScope{}, false, nil
	}
	scope, err := accessScopeFromContext(ctx, workspaceID, fallbackActorID)
	return scope, true, err
}

func (s *Service) authorizeInternalAccess(ctx context.Context, workspaceID, fallbackActorID uuid.UUID) error {
	_, _, err := s.coreScope(ctx, workspaceID, fallbackActorID)
	return err
}

func (s *Service) getInternalItem(ctx context.Context, workspaceID, itemID, fallbackActorID uuid.UUID) (CoreItem, error) {
	scope, scoped, err := s.coreScope(ctx, workspaceID, fallbackActorID)
	if err != nil {
		return CoreItem{}, err
	}
	if scoped {
		return s.scopedCoreRepo.GetItemScoped(ctx, scope, itemID)
	}
	return s.repo.GetItem(ctx, workspaceID, itemID)
}

func (s *Service) getInternalComment(ctx context.Context, workspaceID, itemID, commentID, fallbackActorID uuid.UUID) (CoreComment, error) {
	scope, scoped, err := s.coreScope(ctx, workspaceID, fallbackActorID)
	if err != nil {
		return CoreComment{}, err
	}
	if scoped {
		return s.scopedCoreRepo.GetCommentScoped(ctx, scope, itemID, commentID)
	}
	return s.repo.GetComment(ctx, workspaceID, itemID, commentID)
}

func (s *Service) listInternalStoryLinks(ctx context.Context, workspaceID, itemID, fallbackActorID uuid.UUID) ([]CoreStoryLink, error) {
	scope, scoped, err := s.coreScope(ctx, workspaceID, fallbackActorID)
	if err != nil {
		return nil, err
	}
	if scoped {
		return s.scopedCoreRepo.ListItemStoryLinksScoped(ctx, scope, itemID)
	}
	return s.repo.ListItemStoryLinks(ctx, workspaceID, itemID)
}

func (s *Service) linkInternalStory(ctx context.Context, input CoreStoryLinkInput) (CoreStoryLink, error) {
	scope, scoped, err := s.coreScope(ctx, input.WorkspaceID, input.CreatedByUserID)
	if err != nil {
		return CoreStoryLink{}, err
	}
	if scoped {
		return s.scopedCoreRepo.LinkStoryScoped(ctx, scope, input)
	}
	return s.repo.LinkStory(ctx, input)
}

func (s *Service) markInternalItemRead(ctx context.Context, workspaceID, itemID, actorID uuid.UUID) (time.Time, error) {
	scope, scoped, err := s.coreScope(ctx, workspaceID, actorID)
	if err != nil {
		return time.Time{}, err
	}
	if scoped {
		return s.scopedCoreRepo.MarkItemReadScoped(ctx, scope, itemID)
	}
	return s.repo.MarkItemRead(ctx, workspaceID, itemID, actorID)
}
