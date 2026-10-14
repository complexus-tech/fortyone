"use client";
import { createContext, use, useMemo } from "react";
import type { ReactNode } from "react";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/components/ui/stories-view-options-button";
import type { DetailedStory } from "@/modules/story/types";

type BoardState = {
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
};

type BoardView = Omit<BoardState, "selectedStories" | "setSelectedStories">;

const BoardContext = createContext<BoardState | null>(null);
const BoardViewContext = createContext<BoardView | null>(null);

export const BoardProvider = ({
  children,
  value,
}: {
  children: ReactNode;
  value: BoardState;
}) => {
  const {
    embedded,
    viewOptions,
    setViewOptions,
    isColumnVisible,
    updateStory,
    newStoryDefaults,
  } = value;
  const view = useMemo(
    () => ({
      embedded,
      viewOptions,
      setViewOptions,
      isColumnVisible,
      updateStory,
      newStoryDefaults,
    }),
    [
      embedded,
      viewOptions,
      setViewOptions,
      isColumnVisible,
      updateStory,
      newStoryDefaults,
    ],
  );

  return (
    <BoardContext.Provider value={value}>
      <BoardViewContext.Provider value={view}>
        {children}
      </BoardViewContext.Provider>
    </BoardContext.Provider>
  );
};

// Cards and display controls do not subscribe to the full selection array.
export const useBoardView = () => {
  const context = use(BoardViewContext);
  if (!context) {
    throw new Error("useBoardView must be used within a BoardProvider");
  }
  return context;
};

export const useBoard = () => {
  const context = use(BoardContext);
  if (!context) {
    throw new Error("useBoard must be used within a BoardProvider");
  }
  return context;
};
