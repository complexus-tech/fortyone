import type {
  GroupStoryParams,
  GroupedStoriesResponse,
} from "@/modules/stories/types";

export const groupFilters = (meta: GroupedStoriesResponse["meta"]) => {
  const finalFilters: Omit<GroupStoryParams, "groupKey" | "groupBy"> = {
    teamIds: meta.filters.teamIds ?? undefined,
    sprintIds: meta.filters.sprintIds ?? undefined,
    excludedSprintIds: meta.filters.excludedSprintIds ?? undefined,
    statusIds: meta.filters.statusIds ?? undefined,
    excludedStatusIds: meta.filters.excludedStatusIds ?? undefined,
    assigneeIds: meta.filters.assigneeIds ?? undefined,
    excludedAssigneeIds: meta.filters.excludedAssigneeIds ?? undefined,
    categories: meta.filters.categories ?? undefined,
    orderBy: meta.orderBy,
    orderDirection: meta.orderDirection,
    // page?: number;
    // pageSize?: number;
    assignedToMe: meta.filters.assignedToMe ?? undefined,
    createdByMe: meta.filters.createdByMe ?? undefined,
    reporterIds: meta.filters.reporterIds ?? undefined,
    excludedReporterIds: meta.filters.excludedReporterIds ?? undefined,
    titleContains: meta.filters.titleContains ?? undefined,
    titleNotContains: meta.filters.titleNotContains ?? undefined,
    priorities: meta.filters.priorities ?? undefined,
    excludedPriorities: meta.filters.excludedPriorities ?? undefined,
    excludedTeamIds: meta.filters.excludedTeamIds ?? undefined,
    labelIds: meta.filters.labelIds ?? undefined,
    excludedLabelIds: meta.filters.excludedLabelIds ?? undefined,
    estimateValues: meta.filters.estimateValues ?? undefined,
    excludedEstimateValues: meta.filters.excludedEstimateValues ?? undefined,
    parentId: meta.filters.parentId ?? undefined,
    objectiveId: meta.filters.objectiveId ?? undefined,
    epicId: meta.filters.epicId ?? undefined,
    hasNoAssignee: meta.filters.hasNoAssignee ?? undefined,
    hasBlockedBy: meta.filters.hasBlockedBy ?? undefined,
    createdAfter: meta.filters.createdAfter ?? undefined,
    createdBefore: meta.filters.createdBefore ?? undefined,
    updatedAfter: meta.filters.updatedAfter ?? undefined,
    updatedBefore: meta.filters.updatedBefore ?? undefined,
    startDateAfter: meta.filters.startDateAfter ?? undefined,
    startDateBefore: meta.filters.startDateBefore ?? undefined,
    deadlineAfter: meta.filters.deadlineAfter ?? undefined,
    deadlineBefore: meta.filters.deadlineBefore ?? undefined,
    includeArchived: meta.filters.includeArchived ?? undefined,
    showSubStories: meta.filters.showSubStories ?? undefined,
    completedAfter: meta.filters.completedAfter ?? undefined,
    completedBefore: meta.filters.completedBefore ?? undefined,
    includeDeleted: meta.filters.includeDeleted ?? undefined,
  };
  return finalFilters;
};
