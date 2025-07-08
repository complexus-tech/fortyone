import type {
  GroupStoryParams,
  GroupedStoriesResponse,
} from "@/modules/stories/types";

export const groupFilters = (meta: GroupedStoriesResponse["meta"]) => {
  const finalFilters: Omit<GroupStoryParams, "groupKey" | "groupBy"> = {
    teamIds: meta.filters.teamIds ?? undefined,
    sprintIds: meta.filters.sprintIds ?? undefined,
    statusIds: meta.filters.statusIds ?? undefined,
    assigneeIds: meta.filters.assigneeIds ?? undefined,
    orderBy: meta.orderBy,
    orderDirection: meta.orderDirection,
    // page?: number;
    // pageSize?: number;
    assignedToMe: meta.filters.assignedToMe ?? undefined,
    createdByMe: meta.filters.createdByMe ?? undefined,
    reporterIds: meta.filters.reporterIds ?? undefined,
    priorities: meta.filters.priorities ?? undefined,
    labelIds: meta.filters.labelIds ?? undefined,
    parentId: meta.filters.parentId ?? undefined,
    objectiveId: meta.filters.objectiveId ?? undefined,
    epicId: meta.filters.epicId ?? undefined,
    hasNoAssignee: meta.filters.hasNoAssignee ?? undefined,
    createdAfter: meta.filters.createdAfter ?? undefined,
    createdBefore: meta.filters.createdBefore ?? undefined,
    updatedAfter: meta.filters.updatedAfter ?? undefined,
    updatedBefore: meta.filters.updatedBefore ?? undefined,
    deadlineAfter: meta.filters.deadlineAfter ?? undefined,
    deadlineBefore: meta.filters.deadlineBefore ?? undefined,
  };
  return finalFilters;
};
