import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import type { StoriesLayout } from "@/components/ui/stories-board";
import type { ViewIconKey } from "@/shared/views/metadata";
import type { MyWorkViewScope } from "./my-work-scope";

export type { MyWorkViewScope } from "./my-work-scope";

/** The view's transport contract is independent of how it is saved. */
export type SavedViewConfiguration = {
  version: 1;
  scope?: MyWorkViewScope;
  description?: string;
  icon?: ViewIconKey | null;
  layout: StoriesLayout;
  filters: StoriesFilter;
  viewOptions: StoriesViewOptions;
};

export type SavedViewLoadState = {
  workspaceSlug: string;
  userId: string;
  teamId: string;
  viewId: string | null;
  status: "idle" | "loading" | "ready" | "error" | "unavailable";
};
