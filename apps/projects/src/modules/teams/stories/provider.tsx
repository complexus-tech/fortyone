"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import type { StoriesFilter } from "@/components/ui/stories-filter-button";

type TeamStories = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
};

const TeamStoriesContext = createContext<TeamStories | undefined>(undefined);

export const TeamStoriesProvider = ({ children }: { children: ReactNode }) => {
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
  const initialFilters: StoriesFilter = {
    activeSprints: false,
    assignedToMe: false,
    dueToday: false,
    dueThisWeek: false,
    completed: false,
    startDate: null,
    endDate: null,
    assingee: [],
    createdFrom: null,
    createdTo: null,
    issueType: [],
    labels: [],
    priority: [],
    createdBy: [],
    sprint: [],
    status: [],
    updatedFrom: null,
    updatedTo: null,
  };
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    "teams:stories:view-options",
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
    <TeamStoriesContext.Provider
      value={{
        viewOptions,
        setViewOptions,
        filters,
        setFilters,
        resetFilters,
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
