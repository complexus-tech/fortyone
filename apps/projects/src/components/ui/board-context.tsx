"use client";
import { createContext, use } from "react";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/components/ui/stories-view-options-button";
import type { DetailedStory } from "@/modules/story/types";

export const BoardContext = createContext<{
  embedded?: boolean;
  selectedStories: string[];
  setSelectedStories: (value: string[]) => void;
  viewOptions: StoriesViewOptions;
  setViewOptions?: (value: StoriesViewOptions) => void;
  isColumnVisible: (column: DisplayColumn) => boolean;
  updateStory: (storyId: string, payload: Partial<DetailedStory>) => void;
  newStoryDefaults: {
    teamId?: string;
    objectiveId?: string;
    sprintId?: string;
  };
} | null>(null);

export const useBoard = () => {
  const context = use(BoardContext);
  if (!context) {
    throw new Error("useBoard must be used within a BoardProvider");
  }
  return context;
};
