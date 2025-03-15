"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import type { StoriesFilter } from "@/components/ui/stories-filter-button";

type ObjectiveOptions = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
};

const ObjectiveOptionsContext = createContext<ObjectiveOptions | undefined>(
  undefined,
);

export const ObjectiveOptionsProvider = ({
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
      "Deadline",
      "Created",
      "Updated",
      // "Sprint",
      // "Labels",
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
    "teams:objectives:stories:view-options",
    initialOptions,
  );
  const [filters, setFilters] = useLocalStorage<StoriesFilter>(
    "teams:objectives:stories:filters",
    initialFilters,
  );

  const resetFilters = () => {
    setFilters(initialFilters);
  };
  return (
    <ObjectiveOptionsContext.Provider
      value={{
        viewOptions,
        setViewOptions,
        filters,
        setFilters,
        resetFilters,
      }}
    >
      {children}
    </ObjectiveOptionsContext.Provider>
  );
};

export const useObjectiveOptions = () => {
  const context = useContext(ObjectiveOptionsContext);
  if (!context) {
    throw new Error(
      "useObjectiveStories must be used within a ObjectiveStoriesProvider",
    );
  }
  return context;
};
