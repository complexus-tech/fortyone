"use client";
import { createContext, useContext } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import { useStoriesFilters } from "@/components/ui/stories-filter-state";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import type { StoriesLayout } from "@/components/ui";

type TeamOptions = {
  viewOptions: StoriesViewOptions;
  initialViewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
};

const TeamOptionsContext = createContext<TeamOptions | undefined>(undefined);

export const TeamOptionsProvider = ({
  children,
  layout,
}: {
  children: ReactNode;
  layout: StoriesLayout;
}) => {
  const initialOptions: StoriesViewOptions = {
    groupBy: "priority",
    orderBy: "priority",
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
    `teams:backlog:view-options:${layout}`,
    initialOptions,
  );
  const { filters, resetFilters, setFilters } = useStoriesFilters();

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
    throw new Error("useTeamStories must be used within a TeamStoriesProvider");
  }
  return context;
};
