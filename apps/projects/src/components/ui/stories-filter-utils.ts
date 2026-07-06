import type { StoriesFilter } from "./stories-filter-types";

const ARRAY_FILTER_KEYS = [
  "statusIds",
  "assigneeIds",
  "reporterIds",
  "priorities",
  "teamIds",
  "sprintIds",
  "labelIds",
  "estimateValues",
] as const;

const STRING_FILTER_KEYS = [
  "parentId",
  "objectiveId",
  "epicId",
  "keyResultId",
  "contentContains",
  "startDate",
  "endDate",
] as const;

export const getActiveStoriesFilterCount = (filters: StoriesFilter) => {
  const arrayFilterCount = ARRAY_FILTER_KEYS.reduce((count, key) => {
    const values = filters[key];
    return values && values.length > 0 ? count + 1 : count;
  }, 0);

  const stringFilterCount = STRING_FILTER_KEYS.reduce((count, key) => {
    const value = filters[key];
    return typeof value === "string" && value.trim() ? count + 1 : count;
  }, 0);

  const booleanFilterCount = [
    filters.hasNoAssignee,
    filters.hasBlockedBy,
  ].filter(Boolean).length;

  return arrayFilterCount + stringFilterCount + booleanFilterCount;
};

export const hasActiveStoriesFilters = (filters: StoriesFilter) =>
  getActiveStoriesFilterCount(filters) > 0;
