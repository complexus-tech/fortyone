import type { GroupedStoriesResponse, Story } from "@/modules/stories/types";
import type { DetailedStory } from "@/modules/story/types";

type StoryGroupingFields = Pick<Story, "assigneeId" | "priority" | "statusId">;
type StoryGroupBy = GroupedStoriesResponse["meta"]["groupBy"];

export const getStoryGroupKey = (
  story: StoryGroupingFields,
  groupBy: StoryGroupBy,
) => {
  switch (groupBy) {
    case "status":
      return story.statusId;
    case "priority":
      return story.priority;
    case "assignee":
      return story.assigneeId ?? "null";
    default:
      return null;
  }
};

export const getStoryDropUpdate = (
  story: StoryGroupingFields,
  groupBy: StoryGroupBy,
  targetKey: string,
): Partial<DetailedStory> | null => {
  if (getStoryGroupKey(story, groupBy) === targetKey) return null;

  switch (groupBy) {
    case "status":
      return { statusId: targetKey };
    case "priority":
      return { priority: targetKey as Story["priority"] };
    case "assignee":
      return { assigneeId: targetKey === "null" ? null : targetKey };
    default:
      return null;
  }
};
