"use client";
import { Box } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { TeamOptionsProvider, useTeamOptions } from "./provider";
import { Header } from "./header";
import type { SavedViewsAction } from "./header";
import { AllStories } from "./all-stories";

const ActiveStoriesFilterBar = () => {
  const { filters, resetFilters, setFilters } = useTeamOptions();

  return (
    <StoriesFilterBar
      filters={filters}
      resetFilters={resetFilters}
      setFilters={setFilters}
    />
  );
};

export const ListStories = ({
  renderSavedViews,
}: {
  renderSavedViews?: SavedViewsAction;
}) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:stories:layout",
    "list",
    { initializeWithValue: false },
  );

  return (
    <TeamOptionsProvider layout={layout}>
      <Box className="flex h-full min-h-0 flex-col">
        <Header
          layout={layout}
          renderSavedViews={renderSavedViews}
          setLayout={setLayout}
        />
        <ActiveStoriesFilterBar />
        <Box className="min-h-0 flex-1">
          <AllStories layout={layout} />
        </Box>
      </Box>
    </TeamOptionsProvider>
  );
};
