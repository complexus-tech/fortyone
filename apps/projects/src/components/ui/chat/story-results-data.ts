import type { StoryPriority } from "@/modules/stories/types";

export type StoryResult = {
  id: string;
  priority: StoryPriority;
  statusId: string;
  title: string;
};

export const STORY_RESULTS_PREVIEW_LIMIT = 5;

type StoryResultsOutput = Record<string, unknown> & {
  stories: unknown[];
  success: true;
};

const STORY_PRIORITIES = new Set<StoryPriority>([
  "No Priority",
  "Urgent",
  "High",
  "Medium",
  "Low",
]);

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const asStoryPriority = (value: unknown): StoryPriority =>
  typeof value === "string" && STORY_PRIORITIES.has(value as StoryPriority)
    ? (value as StoryPriority)
    : "No Priority";

const toStoryResult = (value: unknown): StoryResult | null => {
  if (!isRecord(value)) return null;

  const { id, statusId, title } = value;
  if (
    typeof id !== "string" ||
    typeof statusId !== "string" ||
    typeof title !== "string"
  ) {
    return null;
  }

  return {
    id,
    priority: asStoryPriority(value.priority),
    statusId,
    title,
  };
};

export const isStoryResultsOutput = (
  output: unknown,
): output is StoryResultsOutput =>
  isRecord(output) && output.success === true && Array.isArray(output.stories);

export const getStoryResults = (output: unknown): StoryResult[] => {
  if (!isStoryResultsOutput(output)) return [];

  const results: StoryResult[] = [];
  const seenStoryIds = new Set<string>();

  const addStory = (value: unknown) => {
    const story = toStoryResult(value);
    if (!story || seenStoryIds.has(story.id)) return;

    seenStoryIds.add(story.id);
    results.push(story);
  };

  output.stories.forEach((value) => {
    if (isRecord(value) && Array.isArray(value.stories)) {
      value.stories.forEach(addStory);
      return;
    }

    addStory(value);
  });

  return results;
};

export const getVisibleStoryResults = (
  stories: StoryResult[],
  showAll: boolean,
) => (showAll ? stories : stories.slice(0, STORY_RESULTS_PREVIEW_LIMIT));
