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

  const withoutStory = groups.map((g) => {
    const remaining = g.stories.filter((s) => {
      if (s.id === storyId) {
        moved = {
          ...s,
          subStories: s.subStories,
          ...patch,
        } as DetailedStory;
        return false;
      }
      return true;
    });

    if (remaining.length === g.stories.length) {
      return { ...g, stories: remaining };
    }

    return {
      ...g,
      loadedCount: Math.max(0, g.loadedCount - 1),
      stories: remaining,
      totalCount: Math.max(0, g.totalCount - 1),
    };
  });

  if (!moved || !targetKey) return withoutStory;

  return withoutStory.map((g) =>
    g.key === targetKey
      ? {
          ...g,
          loadedCount: g.loadedCount + 1,
          stories: [moved!, ...g.stories],
          totalCount: g.totalCount + 1,
        }
      : g,
  );
};

/**
 * Parse a React-Query key produced by storyKeys.groupStories().
 * Expected shape: ["stories", workspaceSlug, "group", groupKey, params]
 */
export const parseGroupQueryKey = (
  key: readonly unknown[],
): {
  workspaceSlug: string;
  groupKey: string;
  params: Partial<GroupStoryParams>;
} => {
  if (key.length >= 5 && key[2] === "group") {
    const [, workspaceSlug, , groupKey, params] = key;
    return {
      workspaceSlug: workspaceSlug as string,
      groupKey: groupKey as string,
      params: (params ?? {}) as Partial<GroupStoryParams>,
    };
  }

  // Backward-compatible fallback for older key shape.
  const [, , groupKey, params] = key;
  return {
    workspaceSlug: "",
    groupKey: groupKey as string,
    params: (params ?? {}) as Partial<GroupStoryParams>,
  };
};
