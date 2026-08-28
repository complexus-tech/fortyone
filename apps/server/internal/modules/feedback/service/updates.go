package feedback

import (
	"context"
	"github.com/google/uuid"
	"net/url"
	"strings"
	"unicode/utf8"
)

func (s *Service) ListWorkspaceUpdates(ctx context.Context, workspaceID uuid.UUID, page, pageSize int) (CoreUpdatesPage, error) {
	page, pageSize = normalizeUpdatePagination(page, pageSize)
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if err != nil {
			return CoreUpdatesPage{}, err
		}
		return s.scopedNextRepo.ListWorkspaceUpdatesScoped(ctx, scope, page, pageSize)
	}
	return s.nextRepo.ListWorkspaceUpdates(ctx, workspaceID, page, pageSize)
}
func (s *Service) GetWorkspaceUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) (CoreFeedbackUpdate, error) {
	if workspaceID == uuid.Nil || updateID == uuid.Nil {
		return CoreFeedbackUpdate{}, invalidInput("workspace and update ids are required")
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if err != nil {
			return CoreFeedbackUpdate{}, err
		}
		return s.scopedNextRepo.GetWorkspaceUpdateScoped(ctx, scope, updateID)
	}
	return s.nextRepo.GetWorkspaceUpdate(ctx, workspaceID, updateID)
}
func (s *Service) CreateUpdate(ctx context.Context, input CoreUpdateInput) (CoreFeedbackUpdate, error) {
	if err := s.validateUpdateInput(ctx, &input, false); err != nil {
		return CoreFeedbackUpdate{}, err
	}
	return s.nextRepo.CreateUpdate(ctx, input)
}
func (s *Service) UpdateUpdate(ctx context.Context, input CoreUpdateInput) (CoreFeedbackUpdate, error) {
	if err := s.validateUpdateInput(ctx, &input, true); err != nil {
		return CoreFeedbackUpdate{}, err
	}
	return s.nextRepo.UpdateUpdate(ctx, input)
}
func (s *Service) DeleteUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) error {
	if workspaceID == uuid.Nil || updateID == uuid.Nil {
		return invalidInput("workspace and update ids are required")
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if err != nil {
			return err
		}
		return s.scopedNextRepo.DeleteUpdateScoped(ctx, scope, updateID)
	}
	return s.nextRepo.DeleteUpdate(ctx, workspaceID, updateID)
}
func (s *Service) PublishUpdate(ctx context.Context, workspaceID, updateID, actorID uuid.UUID) (CoreFeedbackUpdate, error) {
	if workspaceID == uuid.Nil || updateID == uuid.Nil || actorID == uuid.Nil {
		return CoreFeedbackUpdate{}, invalidInput("workspace, update, and actor ids are required")
	}
	if s.scopedNextRepo != nil {
		scope, scopeErr := accessScopeFromContext(ctx, workspaceID, actorID)
		if scopeErr != nil {
			return CoreFeedbackUpdate{}, scopeErr
		}
		update, _, err := s.scopedNextRepo.PublishUpdateScoped(ctx, scope, updateID)
		return update, err
	}
	update, _, err := s.nextRepo.PublishUpdate(ctx, workspaceID, updateID, actorID)
	if err != nil {
		return CoreFeedbackUpdate{}, err
	}
	return update, nil
}
func (s *Service) UnpublishUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) (CoreFeedbackUpdate, error) {
	if workspaceID == uuid.Nil || updateID == uuid.Nil {
		return CoreFeedbackUpdate{}, invalidInput("workspace and update ids are required")
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if err != nil {
			return CoreFeedbackUpdate{}, err
		}
		return s.scopedNextRepo.UnpublishUpdateScoped(ctx, scope, updateID)
	}
	return s.nextRepo.UnpublishUpdate(ctx, workspaceID, updateID)
}
func (s *Service) ListPublicUpdates(ctx context.Context, portalSlug string, page, pageSize int) (CoreUpdatesPage, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreUpdatesPage{}, err
	}
	page, pageSize = normalizeUpdatePagination(page, pageSize)
	return s.nextRepo.ListPublicUpdates(ctx, portal.ID, page, pageSize)
}
func (s *Service) GetPublicUpdate(ctx context.Context, portalSlug, updateSlug string) (CoreFeedbackUpdate, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreFeedbackUpdate{}, err
	}
	return s.nextRepo.GetPublicUpdate(ctx, portal.ID, normalizeSlug(updateSlug))
}
func (s *Service) validateUpdateInput(ctx context.Context, input *CoreUpdateInput, requireID bool) error {
	if s.nextRepo == nil {
		return ErrFeatureUnavailable
	}
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.ActorID == uuid.Nil || (requireID && input.UpdateID == uuid.Nil) {
		return invalidInput("workspace, portal, update, and actor ids are required")
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, input.WorkspaceID, input.ActorID)
		if err != nil {
			return err
		}
		input.Access = scope
	}
	portal, err := s.repo.GetPortal(ctx, input.WorkspaceID, input.PortalID)
	if err != nil {
		return err
	}
	if portal.WorkspaceID != input.WorkspaceID {
		return ErrNotFound
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	if input.Title == "" || input.Body == "" {
		return invalidInput("update title and body are required")
	}
	if utf8.RuneCountInString(input.Title) > maxUpdateTitleRunes || utf8.RuneCountInString(input.Body) > maxUpdateBodyRunes {
		return invalidInput("feedback update content is too long")
	}
	if input.Summary != nil {
		value := strings.TrimSpace(*input.Summary)
		if value == "" {
			input.Summary = nil
		} else {
			if utf8.RuneCountInString(value) > maxUpdateSummaryRunes {
				return invalidInputf("update summary must be %d characters or fewer", maxUpdateSummaryRunes)
			}
			input.Summary = &value
		}
	}
	if input.CoverImageURL != nil {
		value := strings.TrimSpace(*input.CoverImageURL)
		if value == "" {
			input.CoverImageURL = nil
		} else {
			if value != "" {
				parsed, parseErr := url.ParseRequestURI(value)
				if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" {
					return invalidInput("cover image URL must be an absolute HTTPS URL")
				}
			}
			input.CoverImageURL = &value
		}
	}
	input.Slug = normalizeSlug(input.Slug)
	if input.Slug == "" {
		input.Slug = normalizeSlug(input.Title) + "-" + uuid.NewString()[:8]
	}
	input.ItemIDs = uniqueNonNilUUIDs(input.ItemIDs)
	return nil
}
func normalizeUpdatePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultUpdatesPageSize
	}
	if pageSize > maxUpdatesPageSize {
		pageSize = maxUpdatesPageSize
	}
	return page, pageSize
}
func uniqueNonNilUUIDs(values []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
