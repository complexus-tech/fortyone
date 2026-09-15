import type { MyWorkTab } from "./types";
import { useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Header } from "./components/header";
import { SafeContainer, Tabs, StoriesListSkeleton } from "@/components/ui";
import { StoriesBoard } from "@/modules/stories/components";
import { StoryFiltersSheet } from "@/modules/stories/components/story-filters-sheet";
import { ActiveStoryFilters } from "@/modules/stories/components/active-story-filters";
import {
  buildStoryFilterParams,
  createTeamStoryFilters,
  getTeamStoryFilters,
  getTeamStoryFilterCount,
} from "@/modules/teams/stories/team-story-filters";
import { useMyStoriesGrouped, useViewOptions } from "./hooks";
import { useAuthStore } from "@/store/auth";
import { storyKeys } from "@/constants/keys";

export const MyWork = () => {
  const sessionEpoch = useAuthStore((state) => state.sessionEpoch);
  return <MyWorkContent key={sessionEpoch} />;
};

function MyWorkContent() {
  const client = useQueryClient();
  const [activeTab, setActiveTab] = useState<MyWorkTab>("all");
  const ownerKey = `my-work:${activeTab}`;
  const [storedFilters, setFilters] = useState(() =>
    createTeamStoryFilters(ownerKey),
  );
  const filters = getTeamStoryFilters(ownerKey, storedFilters);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const { viewOptions, setViewOptions, resetViewOptions, isLoaded } =
    useViewOptions();
  const queryOptions = useMemo(
    () => ({
      groupBy: viewOptions.groupBy,
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      ...buildStoryFilterParams(filters),
      ...(activeTab === "all"
        ? { assignedToMe: true, createdByMe: true }
        : activeTab === "assigned"
          ? { assignedToMe: true }
          : { createdByMe: true }),
    }),
    [
      activeTab,
      filters,
      viewOptions.groupBy,
      viewOptions.orderBy,
      viewOptions.orderDirection,
    ],
  );
  const {
    data: groupedStories,
    isPending,
    error,
    refetch,
    isRefetching,
  } = useMyStoriesGrouped(viewOptions.groupBy, queryOptions);
  const filterCount = getTeamStoryFilterCount(filters);
  const openFilters = () => {
    setFilters(filters);
    setFiltersOpen(true);
  };
  return (
    <SafeContainer isFull>
      <Header
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
        onFilters={openFilters}
        filterCount={filterCount}
      />
      {isLoaded ? (
        <Tabs
          value={activeTab}
          onValueChange={(value) => {
            if (value === activeTab) return;
            if (
              value === "all" ||
              value === "assigned" ||
              value === "created"
            ) {
              setActiveTab(value);
              setFilters(createTeamStoryFilters(`my-work:${value}`));
            }
          }}
        >
          <Tabs.List style={{ marginBottom: 4 }}>
            <Tabs.Tab value="all">All</Tabs.Tab>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
          <ActiveStoryFilters
            filters={filters}
            onChange={setFilters}
            allowAssignee={activeTab !== "assigned"}
          />
          <Tabs.Panel value={activeTab}>
            <StoriesBoard
              groupedStories={groupedStories}
              groupFilters={queryOptions}
              isLoading={isPending}
              error={error}
              onRetry={() => {
                void refetch();
              }}
              visibleColumns={viewOptions.displayColumns}
              onRefresh={() => {
                void client.invalidateQueries({ queryKey: storyKeys.all });
              }}
              isRefreshing={isRefetching}
              emptyTitle={filterCount ? "No matching results" : undefined}
              emptyMessage={
                filterCount
                  ? "Try changing or clearing your filters."
                  : undefined
              }
            />
          </Tabs.Panel>
        </Tabs>
      ) : (
        <StoriesListSkeleton />
      )}
      <StoryFiltersSheet
        key={ownerKey}
        isOpen={filtersOpen}
        onClose={() => setFiltersOpen(false)}
        filters={filters}
        onChange={setFilters}
        allowAssignee={activeTab !== "assigned"}
      />
    </SafeContainer>
  );
}
