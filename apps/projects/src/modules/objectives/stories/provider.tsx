"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import type { StoriesFilter } from "@/components/ui/stories-filter-button";
import type { StoriesLayout } from "@/components/ui";

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
    `teams:objectives:stories:view-options:${layout}`,
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
      "useObjectiveOptions must be used within a ObjectiveOptionsProvider",
    );
  }
  return context;
};
