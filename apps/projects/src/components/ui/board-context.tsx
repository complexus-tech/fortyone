"use client";
import { createContext, use } from "react";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/components/ui/stories-view-options-button";

export const BoardContext = createContext<{
  selectedStories: string[];
  setSelectedStories: (value: string[]) => void;
  viewOptions: StoriesViewOptions;
  setViewOptions?: (value: StoriesViewOptions) => void;
  isColumnVisible: (column: DisplayColumn) => boolean;
} | null>(null);

export const useBoard = () => {
  const context = use(BoardContext);
  if (!context) {
    throw new Error("useBoard must be used within a BoardProvider");
  }
  return context;
};
