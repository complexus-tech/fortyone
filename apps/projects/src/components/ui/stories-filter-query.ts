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
  const objectiveOperator = getStoriesFilterOperator(filters, "objectiveId");
  const startDateOperator = getStoriesFilterOperator(filters, "startDate");
  const endDateOperator = getStoriesFilterOperator(filters, "endDate");
  const assigneePresenceOperator = getStoriesFilterOperator(
    filters,
    "hasNoAssignee",
  );

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
    objectiveId:
      objectiveOperator === "is" ? filters.objectiveId ?? undefined : undefined,
    excludedObjectiveId:
      objectiveOperator === "isNot"
        ? filters.objectiveId ?? undefined
        : undefined,
    startDateAfter:
      startDateOperator === "is" || startDateOperator === "isOnOrAfter"
        ? filters.startDate ?? undefined
        : undefined,
    startDateBefore:
      startDateOperator === "is" || startDateOperator === "isOnOrBefore"
        ? filters.startDate ?? undefined
        : undefined,
    startDateNot:
      startDateOperator === "isNot"
        ? filters.startDate ?? undefined
        : undefined,
    deadlineAfter:
      endDateOperator === "is" || endDateOperator === "isOnOrAfter"
        ? filters.endDate ?? undefined
        : undefined,
    deadlineBefore:
      endDateOperator === "is" || endDateOperator === "isOnOrBefore"
        ? filters.endDate ?? undefined
        : undefined,
    deadlineNot:
      endDateOperator === "isNot" ? filters.endDate ?? undefined : undefined,
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
    hasNoAssignee:
      filters.hasNoAssignee && assigneePresenceOperator === "isEmpty"
        ? true
        : undefined,
    hasAssignee:
      filters.hasNoAssignee && assigneePresenceOperator === "isNotEmpty"
        ? true
        : undefined,
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
