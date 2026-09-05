import { formatISO } from "date-fns";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import type { StoryAttentionView } from "@/shared/story/attention";
import type { State } from "@/types/states";

export const getAttentionStoriesFilters = (
  view: StoryAttentionView,
  date: Date,
  statuses: Pick<State, "id" | "category">[],
): StoriesFilter => ({
  ...DEFAULT_STORIES_FILTER,
  endDate: formatISO(date, { representation: "date" }),
  statusIds: statuses
    .filter(
      ({ category }) => category === "completed" || category === "cancelled",
    )
    .map(({ id }) => id),
  operators: {
    endDate: view === "overdue" ? "isOnOrBefore" : "is",
    statusIds: "isNotAnyOf",
  },
});
