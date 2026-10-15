import type { GroupedStoryParams } from "@/modules/stories/types";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import {
  getStoriesFilterStorageKey,
  mergeStoriesFilterDefaults,
} from "@/components/ui/stories-filter-storage";
import { getMyWorkScopeFilterParams } from "@/shared/story/my-work-scope";

const readStored = (key: string): unknown => {
  try {
    return JSON.parse(localStorage.getItem(key) ?? "null") as unknown;
  } catch {
    return null;
  }
};
const isRecord = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === "object" && !Array.isArray(value);

export const myWorkParams = (path: string): GroupedStoryParams => {
  const storedLayout = readStored("my-stories:stories:layout");
  const layout = storedLayout === "list" ? "list" : "kanban";
  const value = readStored(`my-work:view-options:${layout}`);
  const view = isRecord(value) ? value : {};
  const storedFilters = readStored(getStoriesFilterStorageKey(path));
  let filters = DEFAULT_STORIES_FILTER;
  try {
    if (isRecord(storedFilters)) {
      filters = mergeStoriesFilterDefaults(
        storedFilters as Partial<StoriesFilter>,
      );
    }
    return {
      ...getMyWorkScopeFilterParams(filters, { kind: "my-work", tab: "all" }),
      groupBy:
        view.groupBy === "priority" ||
        view.groupBy === "assignee" ||
        view.groupBy === "none"
          ? view.groupBy
          : "status",
      orderBy:
        view.orderBy === "updated" ||
        view.orderBy === "deadline" ||
        view.orderBy === "priority"
          ? view.orderBy
          : "created",
      orderDirection: view.orderDirection === "asc" ? "asc" : "desc",
      showSubStories: view.showSubStories === false ? undefined : true,
    };
  } catch {
    // Invalid old preferences must not prevent startup. The My Work screen
    // remains the owner of its saved filters and performs its normal request.
    return {
      groupBy: "status",
      assignedToMe: true,
      collaboratingWithMe: true,
      createdByMe: true,
      orderBy: "created",
      orderDirection: "desc",
      showSubStories: true,
    };
  }
};
