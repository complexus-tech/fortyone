import type { DetailedStory } from "../../story/types";
import type { Story, StoryGroup, GroupStoryParams } from "../types";

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
    case "assignee": {
      if (!Object.prototype.hasOwnProperty.call(payload, "assigneeId")) {
        return undefined;
      }

      return payload.assigneeId ?? "null";
    }
    default:
      return undefined;
  }
};

const groupFieldByGroup = {
  assignee: "assigneeId",
  priority: "priority",
  status: "statusId",
} as const;

const patchesActiveGroup = (
  groupBy: GroupStoryParams["groupBy"],
  patch: Partial<DetailedStory>,
) => {
  if (groupBy === "none") return false;

  return Object.prototype.hasOwnProperty.call(
    patch,
    groupFieldByGroup[groupBy],
  );
};

export const patchStories = <T extends Story>(
  stories: T[],
  storyId: string,
  patch: Partial<DetailedStory>,
): T[] => {
  if (!stories.some((story) => story.id === storyId)) return stories;

  return stories.map((story) => {
    if (story.id !== storyId) return story;

    return {
      ...story,
      subStories: story.subStories,
      ...patch,
    } as T;
  });
};

const patchStoryInGroups = (
  groups: StoryGroup[],
  storyId: string,
  patch: Partial<DetailedStory>,
) => {
  if (!groups.some((group) => group.stories.some(({ id }) => id === storyId))) {
    return groups;
  }

  const nextGroups = groups.map((group) => {
    const stories = patchStories(group.stories, storyId, patch);
    if (stories === group.stories) return group;

    return { ...group, stories };
  });

  return nextGroups;
};

/**
 * Patch a story in grouped results, moving it only when the update explicitly
 * changes the field that defines the active grouping.
 */
export const updateStoryInGroups = (
  groups: StoryGroup[],
  storyId: string,
  groupBy: GroupStoryParams["groupBy"],
  patch: Partial<DetailedStory>,
) => {
  if (!patchesActiveGroup(groupBy, patch)) {
    return patchStoryInGroups(groups, storyId, patch);
  }

  return moveStoryBetweenGroups(
    groups,
    storyId,
    computeTargetKey(groupBy, patch),
    patch,
  );
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
  if (!groups.some((group) => group.stories.some(({ id }) => id === storyId))) {
    return groups;
  }

  let moved: DetailedStory | undefined;

  const withoutStory = groups.map((g) => {
    if (!g.stories.some((story) => story.id === storyId)) return g;

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
