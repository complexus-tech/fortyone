package stories

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// MyStories returns a list of stories.
func (s *Service) MyStories(ctx context.Context, workspaceId uuid.UUID) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.list")

	var storyList []CoreStoryList
	var err error
	repository, migrated := s.repo.(scopedReadRepository)
	if migrated {
		scope, scopeErr := readScopeFromContext(ctx, workspaceId)
		if scopeErr != nil {
			return nil, scopeErr
		}
		storyList, err = repository.ListMyVisibleStories(ctx, scope)
	} else {
		legacyRepository, supported := s.repo.(legacyMyStoriesRepository)
		if !supported {
			return nil, errors.New("story repository does not support actor story reads")
		}
		storyList, err = legacyRepository.MyStories(ctx, workspaceId)
	}
	if err != nil {
		return nil, err
	}
	if migrated {
		applyStoryListEstimateLabels(storyList)
	} else if err := s.enrichStoryListEstimates(ctx, workspaceId, storyList); err != nil {
		return nil, err
	}
	return storyList, nil
}

// Get returns the story with the specified ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.Get")

	_, migrated := s.repo.(scopedReadRepository)
	story, err := s.getVisibleStory(ctx, id, workspaceId)
	if err != nil {
		return CoreSingleStory{}, err
	}
	if migrated {
		applySingleStoryEstimateLabels(&story)
	} else if err := s.enrichSingleStoryEstimate(ctx, workspaceId, &story); err != nil {
		return CoreSingleStory{}, err
	}

	return story, nil
}

// List returns the actor-visible stories matching typed filters.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID, filters CoreStoryFilters) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.List")

	repository, migrated := s.repo.(scopedReadRepository)
	if !migrated {
		return nil, errors.New("story repository does not support typed visible lists")
	}
	scope, err := readScopeFromContext(ctx, workspaceId)
	if err != nil {
		return nil, err
	}
	storyList, err := repository.ListVisibleStories(ctx, scope, filters)
	if err != nil {
		return nil, err
	}
	applyStoryListEstimateLabels(storyList)
	return storyList, nil
}

// CountInWorkspace returns the count of stories in a workspace.
func (s *Service) CountInWorkspace(ctx context.Context, workspaceId uuid.UUID) (int, error) {
	repository, migrated := s.repo.(scopedReadRepository)
	if !migrated {
		return 0, errors.New("story repository does not support visible counts")
	}
	scope, err := readScopeFromContext(ctx, workspaceId)
	if err != nil {
		return 0, err
	}
	count, err := repository.CountVisibleStories(ctx, scope)
	if err != nil {
		return 0, fmt.Errorf("counting stories in workspace: %w", err)
	}

	return count, nil
}

// ListGroupedStories returns stories grouped by the specified field with limited stories per group
func (s *Service) ListGroupedStories(ctx context.Context, query CoreStoryQuery) ([]CoreStoryGroup, error) {
	repository, migrated := s.repo.(scopedReadRepository)
	if !migrated {
		return nil, errors.New("story repository does not support typed visible groups")
	}
	scope, err := readScopeFromContext(ctx, query.Filters.WorkspaceID)
	if err != nil {
		return nil, err
	}
	groups, err := repository.ListVisibleGroupedStories(ctx, scope, query)
	if err != nil {
		return nil, fmt.Errorf("listing grouped stories: %w", err)
	}
	for i := range groups {
		applyStoryListEstimateLabels(groups[i].Stories)
	}

	return groups, nil
}

// ListGroupStories returns more stories for a specific group (for load more functionality)
func (s *Service) ListGroupStories(ctx context.Context, groupKey string, query CoreStoryQuery) ([]CoreStoryList, bool, error) {
	repository, migrated := s.repo.(scopedReadRepository)
	if !migrated {
		return nil, false, errors.New("story repository does not support typed visible group pages")
	}
	scope, err := readScopeFromContext(ctx, query.Filters.WorkspaceID)
	if err != nil {
		return nil, false, err
	}
	storyList, hasMore, err := repository.ListVisibleGroupStories(ctx, scope, groupKey, query)
	if err != nil {
		return nil, false, fmt.Errorf("listing group stories: %w", err)
	}
	applyStoryListEstimateLabels(storyList)

	return storyList, hasMore, nil
}

// ListByCategory returns stories filtered by category with pagination
func (s *Service) ListByCategory(ctx context.Context, workspaceId, teamId uuid.UUID, category string, page, pageSize int, showSubStories bool) ([]CoreStoryList, bool, error) {
	var storyList []CoreStoryList
	var hasMore bool
	var err error
	repository, migrated := s.repo.(scopedReadRepository)
	if migrated {
		scope, scopeErr := readScopeFromContext(ctx, workspaceId)
		if scopeErr != nil {
			return nil, false, scopeErr
		}
		storyList, hasMore, err = repository.ListVisibleStoriesByCategory(ctx, scope, teamId, category, page, pageSize, showSubStories)
	} else {
		legacyRepository, supported := s.repo.(legacyCategoryStoriesRepository)
		if !supported {
			return nil, false, errors.New("story repository does not support category reads")
		}
		userID, actorErr := auth.GetUserID(ctx)
		if actorErr != nil {
			return nil, false, fmt.Errorf("%w: %v", ErrStoryReadForbidden, actorErr)
		}
		storyList, hasMore, err = legacyRepository.ListByCategory(ctx, workspaceId, userID, teamId, category, page, pageSize, showSubStories)
	}
	if err != nil {
		return nil, false, fmt.Errorf("listing stories by category: %w", err)
	}
	if migrated {
		applyStoryListEstimateLabels(storyList)
	} else if err := s.enrichStoryListEstimates(ctx, workspaceId, storyList); err != nil {
		return nil, false, err
	}

	return storyList, hasMore, nil
}

// QueryByRef returns a story by team code and sequence ID.
func (s *Service) QueryByRef(ctx context.Context, workspaceId uuid.UUID, storyRef string) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.QueryByRef")

	teamCode, sequenceID, err := s.parseStoryRef(storyRef)
	if err != nil {
		return CoreSingleStory{}, err
	}

	var story CoreSingleStory
	repository, migrated := s.repo.(scopedReadRepository)
	if migrated {
		scope, scopeErr := readScopeFromContext(ctx, workspaceId)
		if scopeErr != nil {
			return CoreSingleStory{}, scopeErr
		}
		story, err = repository.QueryVisibleStoryByRef(ctx, scope, teamCode, sequenceID)
	} else {
		legacyRepository, supported := s.repo.(legacyStoryReferenceRepository)
		if !supported {
			return CoreSingleStory{}, errors.New("story repository does not support reference reads")
		}
		story, err = legacyRepository.QueryByRef(ctx, workspaceId, teamCode, sequenceID)
	}
	if err != nil {
		return CoreSingleStory{}, err
	}
	if migrated {
		applySingleStoryEstimateLabels(&story)
	} else if err := s.enrichSingleStoryEstimate(ctx, workspaceId, &story); err != nil {
		return CoreSingleStory{}, err
	}

	return story, nil
}

// parseStoryRef parses a story reference into team code and sequence ID.
func (s *Service) parseStoryRef(storyRef string) (string, int, error) {
	storyRef = strings.ToUpper(strings.ReplaceAll(storyRef, " ", ""))
	storyRef = strings.ReplaceAll(storyRef, "-", "")

	// Split at the transition from letter to digit
	var teamCode, seqStr string
	for i, ch := range storyRef {
		if ch >= '0' && ch <= '9' {
			teamCode = storyRef[:i]
			seqStr = storyRef[i:]
			break
		}
	}

	if teamCode == "" || seqStr == "" {
		return "", 0, fmt.Errorf("%w: %s", ErrInvalidStoryReference, storyRef)
	}

	seqID, err := strconv.Atoi(seqStr)
	if err != nil {
		return "", 0, fmt.Errorf("%w: invalid sequence number in %s", ErrInvalidStoryReference, storyRef)
	}

	return teamCode, seqID, nil
}

func applySingleStoryEstimateLabels(story *CoreSingleStory) {
	story.EstimateLabel = EstimateLabelFromValue(story.EstimateScheme, story.EstimateValue)
	applyStoryListEstimateLabels(story.SubStories)
	for index := range story.Associations {
		applyStoryListEstimateLabel(&story.Associations[index].Story)
	}
}

func applyStoryListEstimateLabels(storyList []CoreStoryList) {
	for index := range storyList {
		applyStoryListEstimateLabel(&storyList[index])
	}
}

func applyStoryListEstimateLabel(story *CoreStoryList) {
	story.EstimateLabel = EstimateLabelFromValue(story.EstimateScheme, story.EstimateValue)
	applyStoryListEstimateLabels(story.SubStories)
}

func (s *Service) enrichSingleStoryEstimate(ctx context.Context, workspaceID uuid.UUID, story *CoreSingleStory) error {
	schemeCache := map[uuid.UUID]string{}
	scheme, err := s.getEstimateSchemeForTeam(ctx, workspaceID, story.Team, schemeCache)
	if err != nil {
		return err
	}

	story.EstimateScheme = scheme
	story.EstimateLabel = EstimateLabelFromValue(scheme, story.EstimateValue)

	for i := range story.SubStories {
		if err := s.enrichStoryListItemEstimate(ctx, workspaceID, &story.SubStories[i], schemeCache); err != nil {
			return err
		}
	}

	for i := range story.Associations {
		if err := s.enrichStoryListItemEstimate(ctx, workspaceID, &story.Associations[i].Story, schemeCache); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) enrichStoryListEstimates(ctx context.Context, workspaceID uuid.UUID, stories []CoreStoryList) error {
	schemeCache := map[uuid.UUID]string{}
	for i := range stories {
		if err := s.enrichStoryListItemEstimate(ctx, workspaceID, &stories[i], schemeCache); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) enrichStoryListItemEstimate(ctx context.Context, workspaceID uuid.UUID, story *CoreStoryList, schemeCache map[uuid.UUID]string) error {
	scheme, err := s.getEstimateSchemeForTeam(ctx, workspaceID, story.Team, schemeCache)
	if err != nil {
		return err
	}

	story.EstimateScheme = scheme
	story.EstimateLabel = EstimateLabelFromValue(scheme, story.EstimateValue)

	for i := range story.SubStories {
		if err := s.enrichStoryListItemEstimate(ctx, workspaceID, &story.SubStories[i], schemeCache); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) getEstimateSchemeForTeam(ctx context.Context, workspaceID, teamID uuid.UUID, schemeCache map[uuid.UUID]string) (string, error) {
	if scheme, ok := schemeCache[teamID]; ok {
		return scheme, nil
	}

	scheme, err := s.getTeamEstimateScheme(ctx, workspaceID, teamID)
	if err != nil {
		return "", err
	}
	schemeCache[teamID] = scheme
	return scheme, nil
}
