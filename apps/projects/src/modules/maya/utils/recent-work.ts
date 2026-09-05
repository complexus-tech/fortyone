import type { StoryActivity } from "@/shared/story/types";
import type { StoryVisit } from "@/shared/story/recent-history";

export const RECENT_WORK_LIMIT = 3;

export const MAYA_WORK_TABS = ["all", "assigned", "created"] as const;
export type MayaWorkTab = (typeof MAYA_WORK_TABS)[number];

export type RecentStory = {
  storyId: string;
  timestamp: string;
  action: "Opened" | "Worked on";
};

export const getRecentStories = (
  visits: StoryVisit[],
  activities: StoryActivity[],
  limit = RECENT_WORK_LIMIT,
): RecentStory[] => {
  const candidates: RecentStory[] = [
    ...visits.map(({ storyId, visitedAt }) => ({
      storyId,
      timestamp: visitedAt,
      action: "Opened" as const,
    })),
    ...activities.map(({ storyId, createdAt }) => ({
      storyId,
      timestamp: createdAt,
      action: "Worked on" as const,
    })),
  ];
  candidates.sort((a, b) => Date.parse(b.timestamp) - Date.parse(a.timestamp));
  const seen = new Set<string>();
  return candidates
    .filter(({ storyId }) => {
      if (seen.has(storyId)) return false;
      seen.add(storyId);
      return true;
    })
    .slice(0, limit);
};
