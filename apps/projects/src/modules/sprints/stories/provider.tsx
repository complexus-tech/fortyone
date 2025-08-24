"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesLayout, StoriesViewOptions } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { StoriesFilter } from "@/components/ui/stories-filter-button";

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
    displayColumns: [
      "ID",
      "Status",
      "Assignee",
      "Priority",
      "Deadline",
      "Created",
      "Updated",
      "Sprint",
      "Labels",
    ],
  };
  const initialFilters: StoriesFilter = {
    statusIds: null,
    assigneeIds: null,
    reporterIds: null,
    priorities: null,
    teamIds: null,
    sprintIds: null,
    labelIds: null,
    parentId: null,
    objectiveId: null,
    epicId: null,
    keyResultId: null,
    hasNoAssignee: null,
    assignedToMe: false,
    createdByMe: false,
  };
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    `sprints:stories:view-options:${layout}`,
    initialOptions,
  );
  const [filters, setFilters] = useLocalStorage<StoriesFilter>(
    "sprints:stories:filters",
    initialFilters,
  );
  const resetFilters = () => {
    setFilters(initialFilters);
  };

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
