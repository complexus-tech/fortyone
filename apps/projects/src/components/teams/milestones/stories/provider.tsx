"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui";
import { useLocalStorage } from "@/hooks";

type MilestoneStories = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
};

const MilestoneStoriesContext = createContext<MilestoneStories | undefined>(
  undefined,
);

export const MilestoneStoriesProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
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
    "milestones:stories:view-options",
    initialOptions,
  );
  return (
    <MilestoneStoriesContext.Provider
      value={{
        viewOptions,
        setViewOptions,
      }}
    >
      {children}
    </MilestoneStoriesContext.Provider>
  );
};

export const useMilestoneStories = () => {
  const context = useContext(MilestoneStoriesContext);
  if (!context) {
    throw new Error(
      "useMilestoneStories must be used within a MilestoneStoriesProvider",
    );
  }
  return context;
};
