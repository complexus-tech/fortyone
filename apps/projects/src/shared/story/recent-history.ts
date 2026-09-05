import { isStoryUuid } from "@/shared/routing/story";

export type StoryVisit = { storyId: string; visitedAt: string };

export const RECENT_STORY_HISTORY_LIMIT = 20;

export const getRecentStoryHistoryKey = (
  userId: string,
  workspaceSlug: string,
) =>
  `recent-stories:v1:${encodeURIComponent(userId)}:${encodeURIComponent(workspaceSlug)}`;

export const parseStoryVisits = (raw: string | null): StoryVisit[] => {
  if (!raw) return [];
  try {
    const value: unknown = JSON.parse(raw);
    if (!Array.isArray(value)) return [];
    return value
      .filter(
        (item: unknown): item is StoryVisit =>
          typeof item === "object" &&
          item !== null &&
          "storyId" in item &&
          typeof item.storyId === "string" &&
          isStoryUuid(item.storyId) &&
          "visitedAt" in item &&
          typeof item.visitedAt === "string" &&
          Number.isFinite(Date.parse(item.visitedAt)),
      )
      .slice(0, RECENT_STORY_HISTORY_LIMIT)
      .map(({ storyId, visitedAt }) => ({ storyId, visitedAt }));
  } catch {
    // A corrupt or older browser record should not prevent opening work.
    return [];
  }
};

export const addStoryVisit = (visits: StoryVisit[], visit: StoryVisit) =>
  [visit, ...visits.filter(({ storyId }) => storyId !== visit.storyId)].slice(
    0,
    RECENT_STORY_HISTORY_LIMIT,
  );
