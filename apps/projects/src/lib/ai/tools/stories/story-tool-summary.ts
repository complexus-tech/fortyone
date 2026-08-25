import type { DetailedStory } from "@/modules/story/types";

export const toStoryToolSummary = ({ id, title }: DetailedStory) => ({
  id,
  title,
});
