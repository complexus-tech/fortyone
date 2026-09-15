import type { StoryPriority } from "@/modules/stories/types";
import type { TeamStoryFilters } from "@/modules/teams/stories/team-story-filters";
import type { StoryFilterFacet } from "./story-filters.types";

export const STORY_FILTER_PRIORITIES = [
  "Urgent",
  "High",
  "Medium",
  "Low",
  "No Priority",
] as const satisfies readonly StoryPriority[];

function isStoryPriority(value: string): value is StoryPriority {
  return STORY_FILTER_PRIORITIES.some((priority) => priority === value);
}

function toggle(values: string[], id: string) {
  return values.includes(id)
    ? values.filter((value) => value !== id)
    : [...values, id];
}

export function changeStoryFilter(
  filters: TeamStoryFilters,
  facet: StoryFilterFacet,
  id: string | null,
): TeamStoryFilters {
  switch (facet) {
    case "status":
      return { ...filters, statusIds: id ? toggle(filters.statusIds, id) : [] };
    case "priority": {
      if (!id) return { ...filters, priorities: [] };
      if (!isStoryPriority(id)) return filters;
      return {
        ...filters,
        priorities: filters.priorities.includes(id)
          ? filters.priorities.filter((value) => value !== id)
          : [...filters.priorities, id],
      };
    }
    case "sprint":
      return { ...filters, sprintIds: id ? toggle(filters.sprintIds, id) : [] };
    case "objective":
      return {
        ...filters,
        objectiveId: id === filters.objectiveId ? null : id,
      };
    case "assignee": {
      if (!id) return { ...filters, assignee: { kind: "any" } };
      if (id === "unassigned")
        return { ...filters, assignee: { kind: "unassigned" } };
      const ids = toggle(
        filters.assignee.kind === "members" ? filters.assignee.ids : [],
        id,
      );
      return {
        ...filters,
        assignee: ids.length ? { kind: "members", ids } : { kind: "any" },
      };
    }
  }
}
