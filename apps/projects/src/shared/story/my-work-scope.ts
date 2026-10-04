import { addDays, formatISO } from "date-fns";
import type { StateCategory } from "@/types/states";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";

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
export const MY_WORK_CATEGORIES = [
  "backlog",
  "unstarted",
  "started",
  "paused",
  "completed",
  "cancelled",
] as const satisfies readonly StateCategory[];
export const ACTIVE_MY_WORK_CATEGORIES = [
  "backlog",
  "unstarted",
  "started",
  "paused",
] as const satisfies readonly StateCategory[];
export type MyWorkViewScope = {
  kind: "my-work";
  tab: MyWorkTab;
  category?: StateCategory | null;
  overdue?: boolean;
  createdAfter?: string | null;
  createdBefore?: string | null;
};
export const getMyWorkDateValue = (date: Date) =>
  formatISO(date, { representation: "date" });

/** Relationship flags are resolved for the current reader by the stories API. */
export const getMyWorkTabFilterParams = (
  tab: MyWorkTab,
  filters: StoriesFilter,
  now = new Date(),
) => {
  const base = getGroupedStoryFilterParams(filters);
  switch (tab) {
    case "today":
      return {
        ...base,
        assignedToMe: true,
        categories: [...ACTIVE_MY_WORK_CATEGORIES],
        deadlineAfter: getMyWorkDateValue(now),
        deadlineBefore: getMyWorkDateValue(now),
      };
    case "upcoming":
      return {
        ...base,
        assignedToMe: true,
        categories: [...ACTIVE_MY_WORK_CATEGORIES],
        deadlineAfter: getMyWorkDateValue(addDays(now, 1)),
        deadlineBefore: getMyWorkDateValue(addDays(now, 7)),
      };
    case "blocked":
      return {
        ...base,
        assignedToMe: true,
        categories: [...ACTIVE_MY_WORK_CATEGORIES],
        hasBlockedBy: true,
      };
    case "assigned":
      return { ...base, assignedToMe: true };
    case "collaborating":
      return { ...base, collaboratingWithMe: true };
    case "created":
      return { ...base, createdByMe: true };
    default:
      return {
        ...base,
        assignedToMe: true,
        collaboratingWithMe: true,
        createdByMe: true,
      };
  }
};

/** A saved personal scope stays relative to the day it is opened, not saved. */
export const getMyWorkScopeFilterParams = (
  filters: StoriesFilter,
  scope: MyWorkViewScope,
  now = new Date(),
) => {
  const tabFilters = getMyWorkTabFilterParams(scope.tab, filters, now);
  const grouped = getGroupedStoryFilterParams(filters);
  const hasEndDate = Boolean(filters.endDate);
  let categories: StateCategory[] | undefined;
  if (scope.overdue) categories = ["started"];
  else if (scope.category) categories = [scope.category];
  const relativeDeadline = scope.overdue
    ? getMyWorkDateValue(now)
    : tabFilters.deadlineBefore;
  return {
    ...tabFilters,
    categories: categories ?? tabFilters.categories,
    createdAfter: scope.createdAfter ?? tabFilters.createdAfter,
    createdBefore: scope.createdBefore ?? tabFilters.createdBefore,
    deadlineAfter: hasEndDate
      ? grouped.deadlineAfter
      : tabFilters.deadlineAfter,
    deadlineBefore: hasEndDate ? grouped.deadlineBefore : relativeDeadline,
    deadlineNot: hasEndDate ? grouped.deadlineNot : tabFilters.deadlineNot,
  };
};

export const hasMyWorkScopeFilters = (scope?: MyWorkViewScope) =>
  Boolean(
    scope &&
      (scope.category ||
        scope.overdue ||
        scope.createdAfter ||
        scope.createdBefore),
  );
