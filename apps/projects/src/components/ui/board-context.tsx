"use client";
import { createContext, useContext } from "react";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/components/ui/stories-view-options-button";

export const BoardContext = createContext<{
  selectedStories: string[];
  setSelectedStories: (value: string[]) => void;
  viewOptions: StoriesViewOptions;
  isColumnVisible: (column: DisplayColumn) => boolean;
} | null>(null);

export const useBoard = () => {
  const context = useContext(BoardContext);
  if (!context) {
    throw new Error("useBoard must be used within a BoardProvider");
  }
  return context;
};
