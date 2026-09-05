import { formatISO } from "date-fns";
import type { StateCategory } from "@/types/states";

export const STORY_ATTENTION_VIEWS = ["overdue", "today"] as const;
export type StoryAttentionView = (typeof STORY_ATTENTION_VIEWS)[number];

export const getStoryAttentionFilters = (
  view: StoryAttentionView,
  date: Date,
) => {
  const today = formatISO(date, { representation: "date" });
  return {
    assignedToMe: true,
    showSubStories: true,
    categories: [
      "backlog",
      "unstarted",
      "started",
      "paused",
    ] as StateCategory[],
    deadlineAfter: view === "today" ? today : undefined,
    deadlineBefore: today,
  };
};
