"use client";
import { createContext, useContext, useState } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import { useStoriesFilters } from "@/components/ui/stories-filter-state";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import type { StoriesLayout } from "@/components/ui";
import type { SavedViewConfiguration } from "@/shared/story/view-configuration";

type TeamOptions = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
  applyView: (configuration: SavedViewConfiguration) => void;
  viewMetadata: Pick<SavedViewConfiguration, "scope" | "icon" | "description">;
};

const TeamOptionsContext = createContext<TeamOptions | undefined>(undefined);

export const TeamOptionsProvider = ({
  children,
  layout,
  viewId,
}: {
  children: ReactNode;
  layout: StoriesLayout;
  viewId?: string | null;
}) => {
  const [savedMetadata, setSavedMetadata] = useState<{
    viewId: string | null | undefined;
    metadata: TeamOptions["viewMetadata"];
  } | null>(null);
  const viewMetadata =
    viewId && savedMetadata?.viewId === viewId ? savedMetadata.metadata : {};
  const initialOptions: StoriesViewOptions = {
    groupBy: "status",
    orderBy: "created",
    orderDirection: "desc",
    showEmptyGroups: true,
    showSubStories: false,
    displayColumns: [
      "ID",
      "Status",
      "Assignee",
      "Estimate",
      "Time needed",
      "Priority",
      "Deadline",
      "Created",
      "Updated",
      "Sprint",
      "Labels",
    ],
  };
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    `teams:stories:view-options:${layout}`,
    initialOptions,
    { initializeWithValue: false },
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
        viewMetadata,
        applyView: (configuration) => {
          setSavedMetadata({
            viewId,
            metadata: {
              scope: configuration.scope,
              icon: configuration.icon,
              description: configuration.description,
            },
          });
          setFilters(configuration.filters);
          if (configuration.layout === layout) {
            setViewOptions(configuration.viewOptions);
          } else {
            // Restore the destination layout before its local preference hook
            // reads it. The current layout keeps its own last-used options.
            localStorage.setItem(
              `teams:stories:view-options:${configuration.layout}`,
              JSON.stringify(configuration.viewOptions),
            );
          }
        },
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
