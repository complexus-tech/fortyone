"use client";

import type { ReactNode } from "react";
import { Box, Tabs, Text } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { SearchEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { StoriesBoard } from "@/components/ui/stories-board";
import type { DisplayColumn } from "@/components/ui/stories-view-options-button";
import { useTerminology } from "@/hooks/use-terminology-display";
import { ListObjectives } from "@/modules/objectives/components/list-objectives";
import type { GroupedStoriesResponse } from "@/modules/stories/types";
import { useSearch } from "@/modules/search/hooks/use-search";
import type { SearchObjective, SearchStory } from "@/modules/search/types";

const SEARCH_DISPLAY_COLUMNS: DisplayColumn[] = [
  "Status",
  "Assignee",
  "Estimate",
  "Time needed",
  "Priority",
  "ID",
  "Deadline",
  "Labels",
  "Objective",
  "Key Result",
  "Sprint",
];

const EMPTY_SEARCH_STORIES: SearchStory[] = [];
const EMPTY_SEARCH_OBJECTIVES: SearchObjective[] = [];

const toGroupedSearchStories = (
  stories: SearchStory[],
): GroupedStoriesResponse => ({
  groups: [
    {
      key: "none",
      totalCount: stories.length,
      stories,
      loadedCount: stories.length,
      hasMore: false,
      nextPage: 1,
    },
  ],
  meta: {
    totalGroups: 1,
    filters: {},
    groupBy: "none",
    orderBy: "priority",
    orderDirection: "desc",
  },
});

const SearchEmptyState = ({ children }: { children: ReactNode }) => (
  <Box className="flex h-[70vh] flex-col items-center justify-center">
    <SearchEmptyIllustration />
    <Text className="mt-6 max-w-md text-center">{children}</Text>
  </Box>
);

export const WorkspaceSearchPage = () => {
  const { getTermDisplay } = useTerminology();
  const [tab, setTab] = useQueryState(
    "type",
    parseAsStringLiteral(["all", "stories", "objectives"]).withDefault(
      "stories",
    ),
  );
  const [query] = useQueryState("query", {
    defaultValue: "",
  });
  const normalizedQuery = query.trim();
  const hasQuery = normalizedQuery.length > 0;
  const { data: results, isFetching } = useSearch({
    query: normalizedQuery,
    type: tab === "all" ? undefined : tab,
  });
  const stories = results?.stories ?? EMPTY_SEARCH_STORIES;
  const objectives = results?.objectives ?? EMPTY_SEARCH_OBJECTIVES;
  const hasResultsResponse = Boolean(results);
  const groupedStories = toGroupedSearchStories(stories);

  return (
    <Box className="h-full">
      <Tabs
        onValueChange={(value) => {
          setTab(value as typeof tab);
        }}
        value={tab}
      >
        <Box className="border-border d sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px]">
          <Tabs.List>
            <Tabs.Tab value="stories">
              {getTermDisplay("storyTerm", {
                variant: "plural",
                capitalize: true,
              })}
            </Tabs.Tab>
            <Tabs.Tab value="objectives">
              {getTermDisplay("objectiveTerm", {
                variant: "plural",
                capitalize: true,
              })}
            </Tabs.Tab>
            <Tabs.Tab value="all">All results</Tabs.Tab>
          </Tabs.List>
        </Box>
        {!hasQuery ? (
          <Box className="flex h-[calc(100%-3.7rem)] items-center justify-center px-6">
            <Box className="flex flex-col items-center">
              <SearchEmptyIllustration />
              <Text className="mt-8 mb-3" fontSize="3xl">
                Search your workspace
              </Text>
              <Text className="max-w-md text-center" color="muted">
                Find {getTermDisplay("storyTerm", { variant: "plural" })} and{" "}
                {getTermDisplay("objectiveTerm", { variant: "plural" })} by
                typing in the search field above.
              </Text>
            </Box>
          </Box>
        ) : null}
        {hasQuery && isFetching ? <BoardSkeleton layout="list" /> : null}
        {hasQuery && !isFetching ? (
          <>
            <Tabs.Panel value="all">
              <StoriesBoard
                groupedStories={groupedStories}
                isInSearch
                layout="list"
                viewOptions={{
                  groupBy: "none",
                  orderBy: "priority",
                  orderDirection: "desc",
                  showEmptyGroups: true,
                  showSubStories: false,
                  displayColumns: SEARCH_DISPLAY_COLUMNS,
                }}
              />
              <ListObjectives isInSearch objectives={objectives} />
              {hasResultsResponse &&
              objectives.length === 0 &&
              stories.length === 0 ? (
                <SearchEmptyState>
                  No results found for your search. Try again with a different
                  query.
                </SearchEmptyState>
              ) : null}
            </Tabs.Panel>
            <Tabs.Panel value="stories">
              <StoriesBoard
                groupedStories={groupedStories}
                isInSearch
                layout="list"
                viewOptions={{
                  groupBy: "none",
                  orderBy: "priority",
                  orderDirection: "desc",
                  showEmptyGroups: true,
                  showSubStories: false,
                  displayColumns: SEARCH_DISPLAY_COLUMNS,
                }}
              />
              {hasResultsResponse && stories.length === 0 ? (
                <SearchEmptyState>
                  No {getTermDisplay("storyTerm", { variant: "plural" })}{" "}
                  matched your search
                </SearchEmptyState>
              ) : null}
            </Tabs.Panel>
            <Tabs.Panel value="objectives">
              <ListObjectives isInSearch objectives={objectives} />
              {hasResultsResponse && objectives.length === 0 ? (
                <SearchEmptyState>
                  No {getTermDisplay("objectiveTerm", { variant: "plural" })}{" "}
                  matched your search
                </SearchEmptyState>
              ) : null}
            </Tabs.Panel>
          </>
        ) : null}
      </Tabs>
    </Box>
  );
};
