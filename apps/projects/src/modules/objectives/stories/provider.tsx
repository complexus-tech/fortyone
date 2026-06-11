"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import { useStoriesFilters } from "@/components/ui/stories-filter-state";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
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
    showSubStories: false,
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
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    `teams:objectives:stories:view-options:${layout}`,
    initialOptions,
  );
  const { filters, resetFilters, setFilters } = useStoriesFilters();

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
