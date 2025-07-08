"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import type { StoriesFilter } from "@/components/ui/stories-filter-button";

type TeamOptions = {
  viewOptions: StoriesViewOptions;
  initialViewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
};

const TeamOptionsContext = createContext<TeamOptions | undefined>(undefined);

export const TeamOptionsProvider = ({ children }: { children: ReactNode }) => {
  const initialOptions: StoriesViewOptions = {
    groupBy: "priority",
    orderBy: "priority",
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
    "teams:archived:view-options-v2",
    initialOptions,
  );
  const [filters, setFilters] = useLocalStorage<StoriesFilter>(
    "teams:archived:filters",
    initialFilters,
  );

  const resetFilters = () => {
    setFilters(initialFilters);
  };
  return (
    <TeamOptionsContext.Provider
      value={{
        viewOptions,
        setViewOptions,
        filters,
        setFilters,
        resetFilters,
        initialViewOptions: initialOptions,
      }}
    >
      {children}
    </TeamOptionsContext.Provider>
  );
};

export const useTeamOptions = () => {
  const context = useContext(TeamOptionsContext);
  if (!context) {
    throw new Error(
      "useArchivedStories must be used within a TeamStoriesProvider",
    );
  }
  return context;
};
