"use client";
import { Box, Tabs, Text } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useState } from "react";
import { useTerminology } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import type { DisplayColumn } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { ListObjectives } from "../objectives/components/list-objectives";
import { Header } from "./components/header";
import { useSearch } from "./hooks/use-search";
import type { SearchQueryParams } from "./types";

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
  const [query, _] = useQueryState("query", {
    defaultValue: "",
  });
  const [searchParams, setSearchParams] = useState<SearchQueryParams>({
    query,
    type: tab === "all" ? undefined : tab,
  });

  const { data: results, isFetching } = useSearch(searchParams);
  const displayColumns: DisplayColumn[] = [
    "Status",
    "Assignee",
    "Priority",
    "ID",
    "Deadline",
    "Labels",
    "Objective",
    "Sprint",
  ];

  const handleSearch = (params: SearchQueryParams) => {
    setSearchParams(params);
  };

  return (
    <>
      <Header onSearch={handleSearch} />
      <Box className="h-[calc(100vh-4rem)]">
        <Tabs
          defaultValue={tab}
          onValueChange={(v) => {
            setTab(v as typeof tab);
            setSearchParams({
              query,
              type: v === "all" ? undefined : (v as typeof tab),
            });
          }}
          value={tab}
        >
          <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100 dark:border-dark-100">
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
          {isFetching ? (
            <BoardSkeleton layout="list" />
          ) : (
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
                    showEmptyGroups: true,
                    displayColumns,
                  }}
                />
                <ListObjectives
                  isInSearch
                  objectives={results?.objectives || []}
                />
                {results?.objectives.length === 0 &&
                  results.stories.length === 0 && (
                    <Box className="flex h-[70vh] items-center justify-center">
                      <Text className="max-w-md text-center">
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
                    showEmptyGroups: true,
                    displayColumns,
                  }}
                />
                {results?.stories.length === 0 && (
                  <Box className="flex h-[70vh] items-center justify-center">
                    <Text>
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
                  <Box className="flex h-[70vh] items-center justify-center">
                    <Text>
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
          )}
        </Tabs>
      </Box>
    </>
  );
};
