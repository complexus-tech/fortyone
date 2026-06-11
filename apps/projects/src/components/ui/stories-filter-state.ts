"use client";

import { useCallback, useMemo } from "react";
import {
  parseAsArrayOf,
  parseAsBoolean,
  parseAsString,
  type UrlKeys,
  useQueryStates,
} from "nuqs";
import type { StoriesFilter } from "./stories-filter-types";
import { DEFAULT_STORIES_FILTER } from "./stories-filter-types";

const storiesFilterParsers = {
  statusIds: parseAsArrayOf(parseAsString),
  assigneeIds: parseAsArrayOf(parseAsString),
  reporterIds: parseAsArrayOf(parseAsString),
  priorities: parseAsArrayOf(parseAsString),
  teamIds: parseAsArrayOf(parseAsString),
  sprintIds: parseAsArrayOf(parseAsString),
  labelIds: parseAsArrayOf(parseAsString),
  parentId: parseAsString,
  objectiveId: parseAsString,
  epicId: parseAsString,
  keyResultId: parseAsString,
  contentContains: parseAsString,
  startDate: parseAsString,
  endDate: parseAsString,
  hasNoAssignee: parseAsBoolean,
  assignedToMe: parseAsBoolean.withDefault(false),
  createdByMe: parseAsBoolean.withDefault(false),
  completedAfter: parseAsString,
  completedBefore: parseAsString,
  isCompleted: parseAsBoolean,
  isNotCompleted: parseAsBoolean,
};

const storiesFilterUrlKeys = {
  statusIds: "status",
  assigneeIds: "assignee",
  reporterIds: "creator",
  priorities: "priority",
  teamIds: "team",
  sprintIds: "sprint",
  labelIds: "label",
  parentId: "parent",
  objectiveId: "objective",
  epicId: "epic",
  keyResultId: "keyResult",
  contentContains: "content",
  startDate: "storyStartDate",
  endDate: "storyDeadline",
  hasNoAssignee: "noAssignee",
  assignedToMe: "assignedToMe",
  createdByMe: "createdByMe",
  completedAfter: "completedAfter",
  completedBefore: "completedBefore",
  isCompleted: "isCompleted",
  isNotCompleted: "isNotCompleted",
} satisfies UrlKeys<typeof storiesFilterParsers>;

export const useStoriesFilters = () => {
  const [queryFilters, setQueryFilters] = useQueryStates(
    storiesFilterParsers,
    {
      urlKeys: storiesFilterUrlKeys,
    },
  );

  const filters = useMemo<StoriesFilter>(
    () => ({
      ...DEFAULT_STORIES_FILTER,
      ...queryFilters,
    }),
    [queryFilters],
  );

  const setFilters = useCallback(
    (value: StoriesFilter) => {
      void setQueryFilters(value);
    },
    [setQueryFilters],
  );

  const resetFilters = useCallback(() => {
    void setQueryFilters(DEFAULT_STORIES_FILTER);
  }, [setQueryFilters]);

  return {
    filters,
    setFilters,
    resetFilters,
  };
};
