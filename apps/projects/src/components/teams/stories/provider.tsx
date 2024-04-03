"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";

type TeamStories = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
};

const TeamStoriesContext = createContext<TeamStories | undefined>(undefined);

export const TeamStoriesProvider = ({ children }: { children: ReactNode }) => {
  const initialOptions: StoriesViewOptions = {
    groupBy: "Status",
    orderBy: "Priority",
    showEmptyGroups: false,
    displayColumns: [
      "Status",
      "Assignee",
      "Priority",
      "Due date",
      "Created",
      "Updated",
      "Milestone",
      "Labels",
    ],
  };
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    "teams:stories:view-options",
    initialOptions,
  );
  return (
    <TeamStoriesContext.Provider
      value={{
        viewOptions,
        setViewOptions,
      }}
    >
      {children}
    </TeamStoriesContext.Provider>
  );
};

export const useTeamStories = () => {
  const context = useContext(TeamStoriesContext);
  if (!context) {
    throw new Error("useTeamStories must be used within a TeamStoriesProvider");
  }
  return context;
};
