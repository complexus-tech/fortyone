"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui";
import { useLocalStorage } from "@/hooks";

type SprintStories = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
};

const SprintStoriesContext = createContext<SprintStories | undefined>(
  undefined,
);

export const SprintStoriesProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const initialOptions: StoriesViewOptions = {
    groupBy: "Status",
    orderBy: "Priority",
    showEmptyGroups: false,
    displayColumns: [
      "ID",
      "Status",
      "Assignee",
      "Priority",
      "Due date",
      "Created",
      "Updated",
      "Sprint",
      "Labels",
    ],
  };
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    "sprints:stories:view-options",
    initialOptions,
  );
  return (
    <SprintStoriesContext.Provider
      value={{
        viewOptions,
        setViewOptions,
      }}
    >
      {children}
    </SprintStoriesContext.Provider>
  );
};

export const useSprintStories = () => {
  const context = useContext(SprintStoriesContext);
  if (!context) {
    throw new Error(
      "useSprintStories must be used within a SprintStoriesProvider",
    );
  }
  return context;
};
