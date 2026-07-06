"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { useStoriesFilters } from "@/components/ui/stories-filter-state";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";

type MyWork = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
};

const MyWorkContext = createContext<MyWork | undefined>(undefined);

export const MyWorkProvider = ({
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
    showSubStories: true,
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
      "Objective",
      "Labels",
    ],
  };
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    `my-work:view-options:${layout}`,
    initialOptions,
  );
  const { filters, resetFilters, setFilters } = useStoriesFilters();

  return (
    <MyWorkContext.Provider
      value={{
        viewOptions,
        setViewOptions,
        filters,
        setFilters,
        resetFilters,
      }}
    >
      {children}
    </MyWorkContext.Provider>
  );
};

export const useMyWork = () => {
  const context = useContext(MyWorkContext);
  if (!context) {
    throw new Error("useMyWork must be used within a MyWorkProvider");
  }
  return context;
};
