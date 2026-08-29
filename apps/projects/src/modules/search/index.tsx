"use client";
import { Box, Tabs, Text } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useTerminology } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import type { DisplayColumn } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { SearchEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { ListObjectives } from "../objectives/components/list-objectives";
import { useSearch } from "./hooks/use-search";

// type?: "all" | "stories" | "objectives";
// query?: string;
// teamId?: string;
// assigneeId?: string;
// labelId?: string;
// statusId?: string;
// priority?: StoryPriority;
// sortBy?: "relevance" | "updated" | "created";
// page?: number;
// pageSize?: number;

export const SearchPage = () => {
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
  const searchParams = {
    query: normalizedQuery,
    type: tab === "all" ? undefined : tab,
  };

  const { data: results, isFetching } = useSearch(searchParams);
  const displayColumns: DisplayColumn[] = [
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

  return (
    <Box className="h-full">
      <Tabs
        onValueChange={(v) => {
          setTab(v as typeof tab);
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
                groupedStories={{
                  groups: [
                    {
                      key: "none",
                      totalCount: results?.stories.length ?? 0,
                      stories: results?.stories || [],
                      loadedCount: results?.stories.length ?? 0,
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
                }}
                isInSearch
                layout="list"
                viewOptions={{
                  groupBy: "none",
                  orderBy: "priority",
                  orderDirection: "desc",
                  showEmptyGroups: true,
                  showSubStories: false,
                  displayColumns,
                }}
              />
              <ListObjectives
                isInSearch
                objectives={results?.objectives || []}
              />
              {results?.objectives.length === 0 &&
                results.stories.length === 0 && (
                  <Box className="flex h-[70vh] flex-col items-center justify-center">
                    <SearchEmptyIllustration />
                    <Text className="mt-6 max-w-md text-center">
                      No results found for your search. Try again with a
                      different query.
                    </Text>
                  </Box>
                )}
            </Tabs.Panel>
            <Tabs.Panel value="stories">
              <StoriesBoard
                groupedStories={{
                  groups: [
                    {
                      key: "none",
                      totalCount: results?.stories.length ?? 0,
                      stories: results?.stories || [],
                      loadedCount: results?.stories.length ?? 0,
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
                }}
                isInSearch
                layout="list"
                viewOptions={{
                  groupBy: "none",
                  orderBy: "priority",
                  orderDirection: "desc",
                  showEmptyGroups: true,
                  showSubStories: false,
                  displayColumns,
                }}
              />
              {results?.stories.length === 0 && (
                <Box className="flex h-[70vh] flex-col items-center justify-center">
                  <SearchEmptyIllustration />
                  <Text className="mt-6">
                    No{" "}
                    {getTermDisplay("storyTerm", {
                      variant: "plural",
                    })}{" "}
                    matched your search
                  </Text>
                </Box>
              )}
            </Tabs.Panel>
            <Tabs.Panel value="objectives">
              <ListObjectives
                isInSearch
                objectives={results?.objectives || []}
              />
              {results?.objectives.length === 0 && (
                <Box className="flex h-[70vh] flex-col items-center justify-center">
                  <SearchEmptyIllustration />
                  <Text className="mt-6">
                    No{" "}
                    {getTermDisplay("objectiveTerm", {
                      variant: "plural",
                    })}{" "}
                    matched your search
                  </Text>
                </Box>
              )}
            </Tabs.Panel>
          </>
        ) : null}
      </Tabs>
    </Box>
  );
};
