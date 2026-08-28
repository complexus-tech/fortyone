package storiesrepository

import (
	"context"
	"fmt"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

const (
	filteredReadModePage    = "page"
	filteredReadModeGrouped = "grouped"
)

func (r *repo) ListVisibleStories(
	ctx context.Context,
	scope storydomain.ReadScope,
	filters storydomain.StoryFilters,
) ([]storydomain.StoryList, error) {
	if err := r.validateFilteredRead(scope); err != nil {
		return nil, err
	}
	filters, err := normalizeStoryFilters(filters)
	if err != nil {
		return nil, err
	}
	limit := maxFilteredStoryResults
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	resultLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, fmt.Errorf("%w: result limit is outside the supported window", storydomain.ErrInvalidReadQuery)
	}
	resultOffset, err := safecast.Int32(filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("%w: result offset is outside the supported window", storydomain.ErrInvalidReadQuery)
	}
	rows, err := r.reads.ListVisibleFilteredStoryRows(ctx, filteredStoryParams(scope, filters, filteredReadOptions{
		groupBy: storydomain.StoryGroupNone, orderBy: storydomain.StoryOrderCreated,
		direction: storydomain.SortDescending, mode: filteredReadModePage,
		limit: resultLimit, offset: resultOffset,
	}))
	if err != nil {
		return nil, fmt.Errorf("list visible stories: %w", err)
	}
	mapped, err := mapFilteredStoryRows(rows)
	if err != nil {
		return nil, err
	}
	stories := filteredStories(mapped)
	if err := r.attachVisibleSubStories(ctx, scope, stories); err != nil {
		return nil, err
	}
	return stories, nil
}

func (r *repo) CountVisibleStories(ctx context.Context, scope storydomain.ReadScope) (int, error) {
	if err := r.validateFilteredRead(scope); err != nil {
		return 0, err
	}
	count, err := r.reads.CountVisibleStories(ctx, storyreadsql.CountVisibleStoriesParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return 0, fmt.Errorf("count visible stories: %w", err)
	}
	return int(count), nil
}

func (r *repo) ListVisibleGroupedStories(
	ctx context.Context,
	scope storydomain.ReadScope,
	query storydomain.StoryQuery,
) ([]storydomain.StoryGroup, error) {
	if err := r.validateFilteredRead(scope); err != nil {
		return nil, err
	}
	query, err := validateInitialStoryGroupQuery(query)
	if err != nil {
		return nil, err
	}
	if query.GroupBy == storydomain.StoryGroupNone {
		return r.listVisibleUngroupedStories(ctx, scope, query)
	}

	keys, err := r.reads.ListVisibleStoryGroupCatalog(ctx, groupCatalogParams(scope, query))
	if err != nil {
		return nil, fmt.Errorf("list visible story groups: %w", err)
	}
	if len(keys) > maxStoryGroupCatalog {
		return nil, fmt.Errorf("%w: group catalog exceeds %d entries", storydomain.ErrInvalidReadQuery, maxStoryGroupCatalog)
	}
	resultLimit, err := safecast.Int64ToInt32(int64(query.StoriesPerGroup) + 1)
	if err != nil {
		return nil, fmt.Errorf("%w: grouped result limit is outside the supported window", storydomain.ErrInvalidReadQuery)
	}
	rows, err := r.reads.ListVisibleFilteredStoryRows(ctx, filteredStoryParams(scope, query.Filters, filteredReadOptions{
		groupBy: query.GroupBy, orderBy: query.OrderBy, direction: query.OrderDirection,
		mode: filteredReadModeGrouped, limit: resultLimit,
	}))
	if err != nil {
		return nil, fmt.Errorf("list visible grouped stories: %w", err)
	}
	mapped, err := mapFilteredStoryRows(rows)
	if err != nil {
		return nil, err
	}
	groups := groupedStoryRows(keys, mapped, query.StoriesPerGroup)
	if err := r.attachGroupedSubStories(ctx, scope, groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *repo) ListVisibleGroupStories(
	ctx context.Context,
	scope storydomain.ReadScope,
	groupKey string,
	query storydomain.StoryQuery,
) ([]storydomain.StoryList, bool, error) {
	if err := r.validateFilteredRead(scope); err != nil {
		return nil, false, err
	}
	query, err := validateStoryQuery(query)
	if err != nil {
		return nil, false, err
	}
	offset, limit, err := validateGroupPage(query, groupKey)
	if err != nil {
		return nil, false, err
	}
	rows, err := r.reads.ListVisibleFilteredStoryRows(ctx, filteredStoryParams(scope, query.Filters, filteredReadOptions{
		groupBy: query.GroupBy, orderBy: query.OrderBy, direction: query.OrderDirection,
		applyGroup: query.GroupBy != storydomain.StoryGroupNone, groupKey: groupKey,
		mode: filteredReadModePage, limit: limit, offset: offset,
	}))
	if err != nil {
		return nil, false, fmt.Errorf("list visible story group page: %w", err)
	}
	mapped, err := mapFilteredStoryRows(rows)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(mapped) > query.PageSize
	if hasMore {
		mapped = mapped[:query.PageSize]
	}
	stories := filteredStories(mapped)
	if err := r.attachVisibleSubStories(ctx, scope, stories); err != nil {
		return nil, false, err
	}
	return stories, hasMore, nil
}

func (r *repo) listVisibleUngroupedStories(
	ctx context.Context,
	scope storydomain.ReadScope,
	query storydomain.StoryQuery,
) ([]storydomain.StoryGroup, error) {
	pageQuery := query
	pageQuery.PageSize = query.StoriesPerGroup
	if pageQuery.Page == 0 {
		pageQuery.Page = 1
	}
	offset, limit, err := validateGroupPage(pageQuery, "none")
	if err != nil {
		return nil, err
	}
	rows, err := r.reads.ListVisibleFilteredStoryRows(ctx, filteredStoryParams(scope, query.Filters, filteredReadOptions{
		groupBy: storydomain.StoryGroupNone, orderBy: query.OrderBy, direction: query.OrderDirection,
		mode: filteredReadModePage, limit: limit, offset: offset,
	}))
	if err != nil {
		return nil, fmt.Errorf("list visible ungrouped stories: %w", err)
	}
	mapped, err := mapFilteredStoryRows(rows)
	if err != nil {
		return nil, err
	}
	totalCount := 0
	if len(mapped) > 0 {
		totalCount = mapped[0].totalCount
	}
	hasMore := len(mapped) > query.StoriesPerGroup
	if hasMore {
		mapped = mapped[:query.StoriesPerGroup]
	}
	stories := filteredStories(mapped)
	if err := r.attachVisibleSubStories(ctx, scope, stories); err != nil {
		return nil, err
	}
	nextPage := 0
	if hasMore {
		nextPage = pageQuery.Page + 1
	}
	return []storydomain.StoryGroup{{
		Key: "none", LoadedCount: len(stories), TotalCount: totalCount,
		HasMore: hasMore, Stories: stories, NextPage: nextPage,
	}}, nil
}

func (r *repo) validateFilteredRead(scope storydomain.ReadScope) error {
	if r.reads == nil {
		return errReadRepositoryNotConfigured
	}
	return validateReadScope(scope)
}

func filteredStories(rows []filteredStoryRow) []storydomain.StoryList {
	stories := make([]storydomain.StoryList, len(rows))
	for index := range rows {
		stories[index] = rows[index].story
	}
	return stories
}

func groupedStoryRows(keys []string, rows []filteredStoryRow, pageSize int) []storydomain.StoryGroup {
	groups := make([]storydomain.StoryGroup, len(keys))
	indexByKey := make(map[string]int, len(keys))
	for index, key := range keys {
		indexByKey[key] = index
		groups[index] = storydomain.StoryGroup{Key: key, Stories: []storydomain.StoryList{}}
	}
	for _, row := range rows {
		index, exists := indexByKey[row.groupKey]
		if !exists {
			continue
		}
		groups[index].TotalCount = row.totalCount
		if len(groups[index].Stories) < pageSize {
			groups[index].Stories = append(groups[index].Stories, row.story)
		}
	}
	for index := range groups {
		groups[index].LoadedCount = len(groups[index].Stories)
		groups[index].HasMore = groups[index].TotalCount > groups[index].LoadedCount
		if groups[index].HasMore {
			groups[index].NextPage = 2
		}
	}
	return groups
}

func (r *repo) attachGroupedSubStories(
	ctx context.Context,
	scope storydomain.ReadScope,
	groups []storydomain.StoryGroup,
) error {
	total := 0
	for index := range groups {
		total += len(groups[index].Stories)
	}
	flat := make([]storydomain.StoryList, 0, total)
	for index := range groups {
		flat = append(flat, groups[index].Stories...)
	}
	if err := r.attachVisibleSubStories(ctx, scope, flat); err != nil {
		return err
	}
	offset := 0
	for index := range groups {
		count := len(groups[index].Stories)
		groups[index].Stories = append([]storydomain.StoryList(nil), flat[offset:offset+count]...)
		offset += count
	}
	return nil
}
