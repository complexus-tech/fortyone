import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import type { StoriesLayout } from "@/components/ui/stories-board";

/** The view's transport contract is independent of how it is saved. */
export type SavedViewConfiguration = {
  version: 1;
  layout: StoriesLayout;
  filters: StoriesFilter;
  viewOptions: StoriesViewOptions;
};
