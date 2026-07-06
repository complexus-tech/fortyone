"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesLayout, StoriesViewOptions } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { useStoriesFilters } from "@/components/ui/stories-filter-state";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";

type SprintStories = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
};

const SprintStoriesContext = createContext<SprintStories | undefined>(
  undefined,
);

export const SprintStoriesProvider = ({
  children,
  layout,
}: {
  children: ReactNode;
  layout: StoriesLayout;
}) => {
  const initialOptions: StoriesViewOptions = {
    groupBy: "status",
    orderBy: "created",
    showEmptyGroups: true,
    showSubStories: false,
    displayColumns: [
      "ID",
      "Status",
      "Assignee",
      "Estimate",
      "Priority",
      "Deadline",
      "Created",
      "Updated",
      "Sprint",
      "Labels",
    ],
  };
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    `sprints:stories:view-options:${layout}`,
    initialOptions,
  );
  const { filters, resetFilters, setFilters } = useStoriesFilters();

  return (
    <SprintStoriesContext.Provider
      value={{
        viewOptions,
        setViewOptions,
        filters,
        setFilters,
        resetFilters,
      }}
    >
      {children}
    </SprintStoriesContext.Provider>
  );
};

export const useSprintOptions = () => {
  const context = useContext(SprintStoriesContext);
  if (!context) {
    throw new Error(
      "useSprintStories must be used within a SprintStoriesProvider",
    );
  }
  return context;
};
