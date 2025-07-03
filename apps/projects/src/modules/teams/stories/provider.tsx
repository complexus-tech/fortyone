"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import type { StoriesFilter } from "@/components/ui/stories-filter-button";

type TeamOptions = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
};

const TeamOptionsContext = createContext<TeamOptions | undefined>(undefined);

export const TeamOptionsProvider = ({ children }: { children: ReactNode }) => {
  const initialOptions: StoriesViewOptions = {
    groupBy: "status",
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
      // "Sprint",
      // "Labels",
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
    "teams:stories:view-options-v2",
    initialOptions,
  );
  const [filters, setFilters] = useLocalStorage<StoriesFilter>(
    "teams:stories:filters",
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
      }}
    >
      {children}
    </TeamOptionsContext.Provider>
  );
};

export const useTeamOptions = () => {
  const context = useContext(TeamOptionsContext);
  if (!context) {
    throw new Error("useTeamStories must be used within a TeamStoriesProvider");
  }
  return context;
};
