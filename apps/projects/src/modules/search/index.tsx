"use client";
import { Box, Tabs } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { ObjectiveIcon, StoryIcon } from "icons";
import { useState } from "react";
import { useTerminology } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
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
    parseAsStringLiteral(["all", "stories", "objectives"]).withDefault("all"),
  );
  const [query, _] = useQueryState("query", {
    defaultValue: "",
  });
  const [searchParams, setSearchParams] = useState<SearchQueryParams>({
    query,
    type: tab === "all" ? undefined : tab,
  });

  const { data: results, isFetching } = useSearch(searchParams);

  const handleSearch = (params: SearchQueryParams) => {
    setSearchParams(params);
  };

  const mappedStories = results?.stories.map((story) => ({
    ...story,
    subStories: [],
    labels: [],
  }));

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
          <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100/60 dark:border-dark-100">
            <Tabs.List>
              <Tabs.Tab value="all">All results</Tabs.Tab>
              <Tabs.Tab
                leftIcon={<StoryIcon className="h-[1.2rem]" />}
                value="stories"
              >
                {getTermDisplay("storyTerm", {
                  variant: "plural",
                  capitalize: true,
                })}
              </Tabs.Tab>
              <Tabs.Tab
                leftIcon={<ObjectiveIcon className="h-[1.1rem]" />}
                value="objectives"
              >
                {getTermDisplay("objectiveTerm", {
                  variant: "plural",
                  capitalize: true,
                })}
              </Tabs.Tab>
            </Tabs.List>
          </Box>
          {isFetching ? (
            <BoardSkeleton layout="list" />
          ) : (
            <>
              <Tabs.Panel value="all">
                <StoriesBoard
                  isInSearch
                  layout="list"
                  stories={mappedStories || []}
                  viewOptions={{
                    groupBy: "None",
                    orderBy: "Priority",
                    showEmptyGroups: true,
                    displayColumns: ["Status", "Assignee", "Priority"],
                  }}
                />
                <ListObjectives
                  isInSearch
                  objectives={results?.objectives || []}
                />
              </Tabs.Panel>
              <Tabs.Panel value="stories">
                <StoriesBoard
                  isInSearch
                  layout="list"
                  stories={mappedStories || []}
                  viewOptions={{
                    groupBy: "None",
                    orderBy: "Priority",
                    showEmptyGroups: true,
                    displayColumns: ["Status", "Assignee", "Priority"],
                  }}
                />
              </Tabs.Panel>
              <Tabs.Panel value="objectives">
                <ListObjectives
                  isInSearch
                  objectives={results?.objectives || []}
                />
              </Tabs.Panel>
            </>
          )}
        </Tabs>
      </Box>
    </>
  );
};
