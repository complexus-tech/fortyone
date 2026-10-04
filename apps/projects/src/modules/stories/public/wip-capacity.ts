import type { LoadWipCapacity } from "@/shared/story/wip-capacity";
import { getGroupedStories } from "../queries/get-grouped-stories";

/** The Stories owner supplies its existing unfiltered grouped request adapter. */
export const loadWipCapacity: LoadWipCapacity = (workspaceSlug, params) =>
  getGroupedStories({ workspaceSlug }, params);
