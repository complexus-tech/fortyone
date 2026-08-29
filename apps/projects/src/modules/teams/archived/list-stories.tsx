"use client";
import { Box } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { TeamOptionsProvider, useTeamOptions } from "./provider";
import { Header } from "./header";
import { AllStories } from "./all-stories";

const ActiveStoriesFilterBar = () => {
  const { filters, resetFilters, setFilters } = useTeamOptions();

  return (
    <StoriesFilterBar
      filters={filters}
      hiddenFields={["teamIds"]}
      resetFilters={resetFilters}
      setFilters={setFilters}
    />
  );
};

export const ListArchivedStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:archived:layout",
    "list",
  );

  return (
    <TeamOptionsProvider layout={layout}>
      <Box className="flex h-full min-h-0 flex-col">
        <Header layout={layout} setLayout={setLayout} />
        <ActiveStoriesFilterBar />
        <Box className="min-h-0 flex-1">
          <AllStories layout={layout} />
        </Box>
      </Box>
    </TeamOptionsProvider>
  );
};
