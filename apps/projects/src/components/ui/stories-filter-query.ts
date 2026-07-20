import type { GroupedStoryParams } from "@/modules/stories/types";
import type { StoriesFilter } from "./stories-filter-types";
import { getStoriesFilterOperator } from "./stories-filter-types";

export const getGroupedStoryFilterParams = (
  filters: StoriesFilter,
): Partial<GroupedStoryParams> => {
  const content = filters.contentContains?.trim() || undefined;
  const isExcluded = (
    field: Parameters<typeof getStoriesFilterOperator>[1],
  ) => {
    const operator = getStoriesFilterOperator(filters, field);
    return operator === "doesNotContain" || operator === "isNotAnyOf";
  };

  return {
    statusIds: isExcluded("statusIds")
      ? undefined
      : filters.statusIds ?? undefined,
    excludedStatusIds: isExcluded("statusIds")
      ? filters.statusIds ?? undefined
      : undefined,
    priorities: isExcluded("priorities")
      ? undefined
      : filters.priorities ?? undefined,
    excludedPriorities: isExcluded("priorities")
      ? filters.priorities ?? undefined
      : undefined,
    assigneeIds: isExcluded("assigneeIds")
      ? undefined
      : filters.assigneeIds ?? undefined,
    excludedAssigneeIds: isExcluded("assigneeIds")
      ? filters.assigneeIds ?? undefined
      : undefined,
    reporterIds: isExcluded("reporterIds")
      ? undefined
      : filters.reporterIds ?? undefined,
    excludedReporterIds: isExcluded("reporterIds")
      ? filters.reporterIds ?? undefined
      : undefined,
    titleContains: isExcluded("contentContains") ? undefined : content,
    titleNotContains: isExcluded("contentContains") ? content : undefined,
    objectiveId: filters.objectiveId ?? undefined,
    startDateAfter: filters.startDate ?? undefined,
    startDateBefore: filters.startDate ?? undefined,
    deadlineAfter: filters.endDate ?? undefined,
    deadlineBefore: filters.endDate ?? undefined,
    teamIds: isExcluded("teamIds") ? undefined : filters.teamIds ?? undefined,
    excludedTeamIds: isExcluded("teamIds")
      ? filters.teamIds ?? undefined
      : undefined,
    sprintIds: isExcluded("sprintIds")
      ? undefined
      : filters.sprintIds ?? undefined,
    excludedSprintIds: isExcluded("sprintIds")
      ? filters.sprintIds ?? undefined
      : undefined,
    labelIds: isExcluded("labelIds")
      ? undefined
      : filters.labelIds ?? undefined,
    excludedLabelIds: isExcluded("labelIds")
      ? filters.labelIds ?? undefined
      : undefined,
    estimateValues: isExcluded("estimateValues")
      ? undefined
      : filters.estimateValues ?? undefined,
    excludedEstimateValues: isExcluded("estimateValues")
      ? filters.estimateValues ?? undefined
      : undefined,
    hasNoAssignee: filters.hasNoAssignee ? true : undefined,
    hasBlockedBy: filters.hasBlockedBy ? true : undefined,
  };
};

export const getScopedStoriesFilterTeamId = (
  routeTeamId: string | undefined,
  selectedTeamIds: string[] | null | undefined,
  teamOperator: ReturnType<typeof getStoriesFilterOperator> = "isAnyOf",
) => {
  if (routeTeamId) {
    return routeTeamId;
  }

  if (teamOperator === "isNotAnyOf") {
    return undefined;
  }

  return selectedTeamIds?.length === 1 ? selectedTeamIds[0] : undefined;
};
