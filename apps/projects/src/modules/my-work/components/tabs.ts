import { addDays, formatISO } from "date-fns";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import type {
  GroupedStoriesResponse,
  GroupedStoryParams,
} from "@/modules/stories/types";
import type { StateCategory } from "@/types/states";

export const MY_WORK_TABS = [
  "all",
  "today",
  "upcoming",
  "blocked",
  "assigned",
  "collaborating",
  "created",
] as const;

export type MyWorkTab = (typeof MY_WORK_TABS)[number];

export const STABLE_MY_WORK_TABS = [
  "all",
  "assigned",
  "collaborating",
  "created",
] as const;

export const ACTIVE_MY_WORK_CATEGORIES = [
  "backlog",
  "unstarted",
  "started",
  "paused",
] as const satisfies readonly StateCategory[];

export const getMyWorkDateValue = (date: Date) =>
  formatISO(date, { representation: "date" });

export const getMyWorkStoriesTotalCount = (
  groupedStories?: GroupedStoriesResponse,
) =>
  groupedStories?.groups.reduce(
    (total, group) => total + group.totalCount,
    0,
  ) ?? 0;

export const getMyWorkTabFilterParams = (
  tab: MyWorkTab,
  filters: StoriesFilter,
): Partial<GroupedStoryParams> => {
  const baseFilters = getGroupedStoryFilterParams(filters);
  const today = getMyWorkDateValue(new Date());
  const tomorrow = getMyWorkDateValue(addDays(new Date(), 1));
  const nextWeek = getMyWorkDateValue(addDays(new Date(), 7));

  switch (tab) {
    case "today":
      return {
        ...baseFilters,
        assignedToMe: true,
        categories: [...ACTIVE_MY_WORK_CATEGORIES],
        deadlineAfter: today,
        deadlineBefore: today,
      };
    case "upcoming":
      return {
        ...baseFilters,
        assignedToMe: true,
        categories: [...ACTIVE_MY_WORK_CATEGORIES],
        deadlineAfter: tomorrow,
        deadlineBefore: nextWeek,
      };
    case "blocked":
      return {
        ...baseFilters,
        assignedToMe: true,
        categories: [...ACTIVE_MY_WORK_CATEGORIES],
        hasBlockedBy: true,
      };
    case "assigned":
      return {
        ...baseFilters,
        assignedToMe: true,
      };
    case "collaborating":
      return {
        ...baseFilters,
        collaboratingWithMe: true,
      };
    case "created":
      return {
        ...baseFilters,
        createdByMe: true,
      };
    case "all":
    default:
      return {
        ...baseFilters,
        assignedToMe: true,
        collaboratingWithMe: true,
        createdByMe: true,
      };
  }
};
