import type { StoriesFilterField } from "@/components/ui/stories-filter-bar";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import { getActiveStoriesFilterCount } from "@/components/ui/stories-filter-utils";

export const PROFILE_HIDDEN_FILTER_FIELDS = [
  "assigneeIds",
  "reporterIds",
  "assignedToMe",
  "createdByMe",
  "hasNoAssignee",
] as const satisfies readonly StoriesFilterField[];

export const hasActiveProfileStoriesFilters = (filters: StoriesFilter) =>
  getActiveStoriesFilterCount({
    ...filters,
    assigneeIds: null,
    reporterIds: null,
    assignedToMe: false,
    createdByMe: false,
    hasNoAssignee: null,
  }) > 0;
