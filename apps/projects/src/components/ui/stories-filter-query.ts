import type { GroupedStoryParams } from "@/modules/stories/types";
import type { StoriesFilter } from "./stories-filter-button";

export const getGroupedStoryFilterParams = (
  filters: StoriesFilter,
): Partial<GroupedStoryParams> => ({
  statusIds: filters.statusIds ?? undefined,
  priorities: filters.priorities ?? undefined,
  assigneeIds: filters.assigneeIds ?? undefined,
  reporterIds: filters.reporterIds ?? undefined,
  titleContains: filters.titleContains?.trim() || undefined,
  objectiveId: filters.objectiveId ?? undefined,
  startDateAfter: filters.startDate ?? undefined,
  startDateBefore: filters.startDate ?? undefined,
  deadlineAfter: filters.endDate ?? undefined,
  deadlineBefore: filters.endDate ?? undefined,
  teamIds: filters.teamIds ?? undefined,
  sprintIds: filters.sprintIds ?? undefined,
  hasNoAssignee: filters.hasNoAssignee ? true : undefined,
});
