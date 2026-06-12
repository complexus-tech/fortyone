import type { StoriesFilter } from "./stories-filter-types";
import { DEFAULT_STORIES_FILTER } from "./stories-filter-types";

const STORIES_FILTER_STORAGE_PREFIX = "stories:filters";

export const getStoriesFilterStorageKey = (pathname: string | null) =>
  `${STORIES_FILTER_STORAGE_PREFIX}:${pathname || "default"}`;

export const mergeStoriesFilterDefaults = (
  value: Partial<StoriesFilter> | null | undefined,
): StoriesFilter => ({
  ...DEFAULT_STORIES_FILTER,
  ...value,
});
