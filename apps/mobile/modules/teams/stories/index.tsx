import type { TeamStoriesTab } from "../types";
import type { TeamStoryFilters } from "./team-story-filters";
import type { StoryFilterFacet } from "@/modules/stories/components/story-filters.types";
import { useMemo, useState } from "react";
import { ScrollView, View } from "react-native";
import { useLocalSearchParams, useRouter } from "expo-router";
import { useQueryClient } from "@tanstack/react-query";
import { Header } from "./components";
import {
  Back,
  SafeContainer,
  Tabs,
  StoriesListSkeleton,
} from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { StoriesBoard } from "@/modules/stories/components";
import { StoryFiltersSheet } from "@/modules/stories/components/story-filters-sheet";
import { ActiveStoryFilters } from "@/modules/stories/components/active-story-filters";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks";
import { useViewOptions } from "@/hooks/use-view-options";
import { useTerminology } from "@/hooks/use-terminology";
import { useAuthStore } from "@/store/auth";
import { storyKeys } from "@/constants/keys";
import {
  buildTeamStoryQuery,
  getTeamStoryFilterCount,
} from "./team-story-filters";
import { parseTeamStoryRouteFilters } from "./team-story-route-filters";

export const TeamStories = () => {
  const params = useLocalSearchParams<{
    teamId: string;
    sprintId?: string | string[];
    objectiveId?: string | string[];
    openFilter?: string | string[];
  }>();
  const sessionEpoch = useAuthStore((state) => state.sessionEpoch);
  const router = useRouter();
  const parsed = parseTeamStoryRouteFilters(params.teamId, params);
  if (parsed.error)
    return (
      <SafeContainer isFull>
        <View style={{ paddingHorizontal: 20 }}>
          <Back />
        </View>
        <QueryState
          title="This filter link is invalid"
          message={parsed.error}
          onRetry={() => router.back()}
          retryLabel="Go back"
        />
      </SafeContainer>
    );
  return (
    <TeamStoriesContent
      key={JSON.stringify([
        sessionEpoch,
        params.teamId,
        params.sprintId,
        params.objectiveId,
        params.openFilter,
      ])}
      teamId={params.teamId}
      initialFilters={parsed.filters}
      initialFacet={parsed.initialFacet}
    />
  );
};

function TeamStoriesContent({
  teamId,
  initialFilters,
  initialFacet,
}: {
  teamId: string;
  initialFilters: TeamStoryFilters;
  initialFacet?: StoryFilterFacet;
}) {
  const client = useQueryClient();
  const [activeTab, setActiveTab] = useState<TeamStoriesTab>("all");
  const [filters, setFilters] = useState(initialFilters);
  const [filtersOpen, setFiltersOpen] = useState(Boolean(initialFacet));
  const [openingFacet, setOpeningFacet] = useState(initialFacet);
  const { viewOptions, setViewOptions, resetViewOptions, isLoaded } =
    useViewOptions(`team-${teamId}:view-options`);
  const { getTermDisplay } = useTerminology();
  const queryOptions = useMemo(
    () => buildTeamStoryQuery({ teamId, filters, tab: activeTab, viewOptions }),
    [teamId, filters, activeTab, viewOptions],
  );
  const {
    data: groupedStories,
    isPending,
    error,
    refetch,
    isRefetching,
  } = useTeamStoriesGrouped(teamId, viewOptions.groupBy, queryOptions);
  const filterCount = getTeamStoryFilterCount(filters);
  const openFilters = () => {
    setOpeningFacet(undefined);
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
            if (value === "all" || value === "active" || value === "backlog")
              setActiveTab(value);
          }}
        >
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            style={{ flexGrow: 0 }}
          >
            <Tabs.List style={{ flexWrap: "nowrap" }}>
              <Tabs.Tab value="all">
                All {getTermDisplay("storyTerm", { variant: "plural" })}
              </Tabs.Tab>
              <Tabs.Tab value="active">In progress</Tabs.Tab>
              <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
            </Tabs.List>
          </ScrollView>
          <ActiveStoryFilters
            filters={filters}
            onChange={setFilters}
            teamId={teamId}
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
              emptyTitle={
                filterCount
                  ? "No matching results"
                  : activeTab === "all"
                    ? `No ${getTermDisplay("storyTerm", { variant: "plural" })} in this team`
                    : activeTab === "active"
                      ? "Nothing in progress"
                      : "The backlog is clear"
              }
              emptyMessage={
                filterCount
                  ? "Try changing or clearing your filters."
                  : "Your team’s work will appear here."
              }
            />
          </Tabs.Panel>
        </Tabs>
      ) : (
        <StoriesListSkeleton />
      )}
      <StoryFiltersSheet
        isOpen={filtersOpen}
        onClose={() => setFiltersOpen(false)}
        filters={filters}
        onChange={setFilters}
        teamId={teamId}
        initialFacet={openingFacet}
      />
    </SafeContainer>
  );
}
