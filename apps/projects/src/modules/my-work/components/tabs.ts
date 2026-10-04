import type { GroupedStoriesResponse } from "@/modules/stories/types";

export {
  MY_WORK_TABS,
  ACTIVE_MY_WORK_CATEGORIES,
  getMyWorkDateValue,
  getMyWorkTabFilterParams,
} from "@/shared/story/my-work-scope";
export type { MyWorkTab } from "@/shared/story/my-work-scope";

export const STABLE_MY_WORK_TABS = [
  "all",
  "assigned",
  "collaborating",
  "created",
] as const;

export const getMyWorkStoriesTotalCount = (
  groupedStories?: GroupedStoriesResponse,
) =>
  groupedStories?.groups.reduce(
    (total, group) => total + group.totalCount,
    0,
  ) ?? 0;
