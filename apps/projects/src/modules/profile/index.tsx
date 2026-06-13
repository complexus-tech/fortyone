"use client";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { Header } from "./components/header";
import { AllStories } from "./components/all-stories";
import { ProfileProvider, useProfile } from "./components/provider";

const ActiveStoriesFilterBar = () => {
  const { filters, resetFilters, setFilters } = useProfile();

  return (
    <StoriesFilterBar
      filters={filters}
      resetFilters={resetFilters}
      setFilters={setFilters}
      showWhenEmpty
    />
  );
};

export const ListUserStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "profile:stories:layout",
    "kanban",
  );
  return (
    <ProfileProvider layout={layout}>
      <Header layout={layout} setLayout={setLayout} />
      <ActiveStoriesFilterBar />
      <AllStories layout={layout} />
    </ProfileProvider>
  );
};
