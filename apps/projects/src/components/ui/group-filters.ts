import type { GroupStoryParams, StoryFilters } from "@/modules/stories/types";

export const groupFilters = (filters: StoryFilters) => {
  const finalFilters: Omit<GroupStoryParams, "groupKey" | "groupBy"> = {
    teamIds: filters.teamIds ?? undefined,
    sprintIds: filters.sprintIds ?? undefined,
    statusIds: filters.statusIds ?? undefined,
    assigneeIds: filters.assigneeIds ?? undefined,
    // orderBy: filters.
    // orderDirection?: "asc" | "desc";
    // page?: number;
    // pageSize?: number;
    assignedToMe: filters.assignedToMe ?? undefined,
    createdByMe: filters.createdByMe ?? undefined,
    reporterIds: filters.reporterIds ?? undefined,
    priorities: filters.priorities ?? undefined,
    labelIds: filters.labelIds ?? undefined,
    parentId: filters.parentId ?? undefined,
    objectiveId: filters.objectiveId ?? undefined,
    epicId: filters.epicId ?? undefined,
    hasNoAssignee: filters.hasNoAssignee ?? undefined,
  };
  return finalFilters;
};
