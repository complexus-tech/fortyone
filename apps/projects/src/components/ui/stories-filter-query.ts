import type { GroupedStoryParams } from "@/modules/stories/types";
import type { StoriesFilter } from "./stories-filter-types";

export const getGroupedStoryFilterParams = (
  filters: StoriesFilter,
): Partial<GroupedStoryParams> => ({
  statusIds: filters.statusIds ?? undefined,
  priorities: filters.priorities ?? undefined,
  assigneeIds: filters.assigneeIds ?? undefined,
  reporterIds: filters.reporterIds ?? undefined,
  titleContains: filters.contentContains?.trim() || undefined,
  objectiveId: filters.objectiveId ?? undefined,
  startDateAfter: filters.startDate ?? undefined,
  startDateBefore: filters.startDate ?? undefined,
  deadlineAfter: filters.endDate ?? undefined,
  deadlineBefore: filters.endDate ?? undefined,
  teamIds: filters.teamIds ?? undefined,
  sprintIds: filters.sprintIds ?? undefined,
  labelIds: filters.labelIds ?? undefined,
  estimateValues: filters.estimateValues ?? undefined,
  hasNoAssignee: filters.hasNoAssignee ? true : undefined,
});

export const getScopedStoriesFilterTeamId = (
  routeTeamId: string | undefined,
  selectedTeamIds: string[] | null | undefined,
) => {
  if (routeTeamId) {
    return routeTeamId;
  }

  return selectedTeamIds?.length === 1 ? selectedTeamIds[0] : undefined;
};
