// New file with helper utilities for optimistic updates
import type { DetailedStory } from "../../story/types";
import type { StoryGroup, GroupStoryParams } from "../types";

/**
 * Compute the group key a story should belong to after an update.
 */
export const computeTargetKey = (
  groupBy: GroupStoryParams["groupBy"],
  payload: Partial<DetailedStory>,
): string | undefined => {
  switch (groupBy) {
    case "status":
      return payload.statusId;
    case "priority":
      return payload.priority as string | undefined;
    case "assignee":
      return payload.assigneeId ?? undefined;
    default:
      return undefined;
  }
};

/**
 * Move a story between StoryGroups in memory.
 * Removes the story from any existing group and inserts it at the start of the
 * target group if provided.
 */
export const moveStoryBetweenGroups = (
  groups: StoryGroup[],
  storyId: string,
  targetKey: string | undefined,
  patch: Partial<DetailedStory>,
): StoryGroup[] => {
  let moved: DetailedStory | undefined;

  // 1. Remove the story from its current group, remembering it.
  const withoutStory = groups.map((g) => {
    const remaining = g.stories.filter((s) => {
      if (s.id === storyId) {
        moved = { ...s, ...patch } as DetailedStory;
        return false;
      }
      return true;
    });
    return { ...g, stories: remaining };
  });

  // If we couldn't find the story or no target, just return cleaned groups.
  if (!moved || !targetKey) return withoutStory;

  // 2. Insert into target group.
  return withoutStory.map((g) =>
    g.key === targetKey ? { ...g, stories: [moved!, ...g.stories] } : g,
  );
};

/**
 * Parse a React-Query key produced by storyKeys.groupStories().
 * Expected shape: ["stories", "group", groupKey, params]
 */
export const parseGroupQueryKey = (
  key: readonly unknown[],
): { groupKey: string; params: Partial<GroupStoryParams> } => {
  const [, , groupKey, params] = key;
  return {
    groupKey: groupKey as string,
    params: (params ?? {}) as Partial<GroupStoryParams>,
  };
};
