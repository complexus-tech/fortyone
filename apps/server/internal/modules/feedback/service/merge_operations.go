package feedback

import (
	"context"
	"github.com/google/uuid"
	"strings"
	"unicode/utf8"
)

func (s *Service) MergeItems(ctx context.Context, input CoreMergeItemInput) (CoreMergeItemResult, error) {
	if s.nextRepo == nil {
		return CoreMergeItemResult{}, ErrFeatureUnavailable
	}
	if input.WorkspaceID == uuid.Nil || input.SourceItemID == uuid.Nil || input.TargetItemID == uuid.Nil || input.ActorID == uuid.Nil {
		return CoreMergeItemResult{}, invalidInput("workspace, source, target, and actor ids are required")
	}
	if input.SourceItemID == input.TargetItemID {
		return CoreMergeItemResult{}, ErrMergeConflict
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, input.WorkspaceID, input.ActorID)
		if err != nil {
			return CoreMergeItemResult{}, err
		}
		input.Access = scope
		return s.scopedNextRepo.MergeItemsScoped(ctx, scope, input)
	}
	return s.nextRepo.MergeItems(ctx, input)
}
func (s *Service) ListMergeCandidates(ctx context.Context, workspaceID, sourceItemID uuid.UUID, search string, limit int) (CoreMergeCandidatesPage, error) {
	if s.nextRepo == nil {
		return CoreMergeCandidatesPage{}, ErrFeatureUnavailable
	}
	if workspaceID == uuid.Nil || sourceItemID == uuid.Nil {
		return CoreMergeCandidatesPage{}, invalidInput("workspace and source feedback ids are required")
	}
	search = strings.TrimSpace(search)
	if utf8.RuneCountInString(search) > maxPublicFeedbackTitleCharacters {
		return CoreMergeCandidatesPage{}, invalidInput("merge candidate search is too long")
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > maxMergeCandidatesLimit {
		limit = maxMergeCandidatesLimit
	}
	source, err := s.getInternalItem(ctx, workspaceID, sourceItemID, uuid.Nil)
	if err != nil {
		return CoreMergeCandidatesPage{}, err
	}
	if source.DeletedAt != nil || source.MergedIntoItemID != nil {
		return CoreMergeCandidatesPage{}, ErrMergeConflict
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if err != nil {
			return CoreMergeCandidatesPage{}, err
		}
		return s.scopedNextRepo.ListItemCandidatesScoped(ctx, scope, source.PortalID, sourceItemID, search, limit)
	}
	return s.nextRepo.ListItemCandidates(ctx, workspaceID, source.PortalID, sourceItemID, search, limit)
}
func (s *Service) ListPortalItemCandidates(ctx context.Context, workspaceID, portalID uuid.UUID, search string, limit int) (CoreMergeCandidatesPage, error) {
	if s.nextRepo == nil {
		return CoreMergeCandidatesPage{}, ErrFeatureUnavailable
	}
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return CoreMergeCandidatesPage{}, invalidInput("workspace and portal ids are required")
	}
	search = strings.TrimSpace(search)
	if utf8.RuneCountInString(search) > maxPublicFeedbackTitleCharacters {
		return CoreMergeCandidatesPage{}, invalidInput("item candidate search is too long")
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > maxMergeCandidatesLimit {
		limit = maxMergeCandidatesLimit
	}
	if _, err := s.repo.GetPortal(ctx, workspaceID, portalID); err != nil {
		return CoreMergeCandidatesPage{}, err
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if err != nil {
			return CoreMergeCandidatesPage{}, err
		}
		return s.scopedNextRepo.ListItemCandidatesScoped(ctx, scope, portalID, uuid.Nil, search, limit)
	}
	return s.nextRepo.ListItemCandidates(ctx, workspaceID, portalID, uuid.Nil, search, limit)
}
