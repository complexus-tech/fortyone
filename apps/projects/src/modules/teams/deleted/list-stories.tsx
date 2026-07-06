"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
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

export const ListDeletedStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:deleted:layout",
    "list",
  );

  return (
    <TeamOptionsProvider layout={layout}>
      <Header layout={layout} setLayout={setLayout} />
      <ActiveStoriesFilterBar />
      <AllStories layout={layout} />
    </TeamOptionsProvider>
  );
};
