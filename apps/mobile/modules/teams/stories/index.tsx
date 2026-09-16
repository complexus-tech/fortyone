import type { TeamSection } from "../sections/team-tabs";
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
  Button,
  Text,
} from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { StoriesBoard } from "@/modules/stories/components";
import { StoryFiltersSheet } from "@/modules/stories/components/story-filters-sheet";
import { ActiveStoryFilters } from "@/modules/stories/components/active-story-filters";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks";
import { useViewOptions } from "@/hooks/use-view-options";
import { useTerminology } from "@/hooks/use-terminology";
import { useAuthStore } from "@/store/auth";
import { TeamObjectives } from "@/modules/objectives";
import { TeamFeed } from "../sections/team-feed";
import { useTeamSections } from "../sections/hooks";
import { getTeamTabs, resolveTeamTab } from "../sections/team-tabs";
import { storyKeys } from "@/constants/keys";
import {
  buildTeamStoryQuery,
  getTeamStoryFilterCount,
} from "./team-story-filters";
import { parseTeamStoryRouteFilters } from "./team-story-route-filters";

export const TeamStories = () => {
  const params = useLocalSearchParams<{
    teamId: string;
    section?: string;
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
        params.section,
      ])}
      teamId={params.teamId}
      initialSection={params.section}
      initialFilters={parsed.filters}
      initialFacet={parsed.initialFacet}
    />
  );
};

function TeamStoriesContent({
  teamId,
  initialFilters,
  initialFacet,
  initialSection = "all",
}: {
  teamId: string;
  initialFilters: TeamStoryFilters;
  initialFacet?: StoryFilterFacet;
  initialSection?: string;
}) {
  const client = useQueryClient();
  const [requestedTab, setActiveTab] = useState<string>(initialSection);
  const [filters, setFilters] = useState(initialFilters);
  const [filtersOpen, setFiltersOpen] = useState(Boolean(initialFacet));
  const [openingFacet, setOpeningFacet] = useState(initialFacet);
  const { viewOptions, setViewOptions, resetViewOptions, isLoaded } =
    useViewOptions(`team-${teamId}:view-options`);
  const { getTermDisplay } = useTerminology();
  const sections = useTeamSections(teamId);
  const tabs = getTeamTabs(
    getTermDisplay("storyTerm", { variant: "plural" }),
    getTermDisplay("objectiveTerm", { variant: "plural", capitalize: true }),
    sections.available,
  );
  const activeTab: TeamSection = resolveTeamTab(requestedTab, tabs);
  const queryOptions = useMemo(
    () => buildTeamStoryQuery({ teamId, filters, tab: "all", viewOptions }),
    [teamId, filters, viewOptions],
  );
  const {
    data: groupedStories,
    isPending,
    error,
    refetch,
    isRefetching,
  } = useTeamStoriesGrouped(
    teamId,
    viewOptions.groupBy,
    queryOptions,
    activeTab === "all",
  );
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
        showStoryActions={activeTab === "all"}
      />
      {isLoaded ? (
        <Tabs
          value={activeTab}
          onValueChange={(value) => {
            setActiveTab(resolveTeamTab(value, tabs));
          }}
        >
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            style={{ flexGrow: 0 }}
          >
            <Tabs.List accessibilityLabel="Team sections" options={tabs} />
          </ScrollView>
          {sections.error ? (
            <View style={{ paddingHorizontal: 20, paddingVertical: 8, gap: 4 }}>
              <Text fontSize="sm" color="muted">
                Some team sections could not be loaded.
              </Text>
              <Button
                color="tertiary"
                fullWidth={false}
                onPress={sections.retry}
              >
                Retry sections
              </Button>
            </View>
          ) : null}
          <Tabs.Panel value="all">
            <ActiveStoryFilters
              filters={filters}
              onChange={setFilters}
              teamId={teamId}
            />
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
                  : `No ${getTermDisplay("storyTerm", { variant: "plural" })} in this team`
              }
              emptyMessage={
                filterCount
                  ? "Try changing or clearing your filters."
                  : "Your team’s work will appear here."
              }
            />
          </Tabs.Panel>
          <Tabs.Panel value="objectives">
            <TeamObjectives teamId={teamId} />
          </Tabs.Panel>
          <Tabs.Panel value="feedback">
            <TeamFeed teamId={teamId} kind="feedback" />
          </Tabs.Panel>
          <Tabs.Panel value="intake">
            <TeamFeed teamId={teamId} kind="intake" />
          </Tabs.Panel>
        </Tabs>
      ) : (
        <StoriesListSkeleton />
      )}
      {activeTab === "all" ? (
        <StoryFiltersSheet
          isOpen={filtersOpen}
          onClose={() => setFiltersOpen(false)}
          filters={filters}
          onChange={setFilters}
          teamId={teamId}
          initialFacet={openingFacet}
        />
      ) : null}
    </SafeContainer>
  );
}
