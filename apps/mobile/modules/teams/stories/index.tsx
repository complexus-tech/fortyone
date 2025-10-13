import React, { useState, useMemo } from "react";
import { Header } from "./components";
import { SafeContainer, Tabs, StoriesListSkeleton } from "@/components/ui";
import { StoriesBoard } from "@/modules/stories/components";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks";
import { useViewOptions } from "@/hooks/use-view-options";
import { useGlobalSearchParams } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";
import type { TeamStoriesTab } from "../types";

export const TeamStories = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const [activeTab, setActiveTab] = useState<TeamStoriesTab>("all");
  const {
    viewOptions,
    setViewOptions,
    resetViewOptions,
    isLoaded: viewOptionsLoaded,
  } = useViewOptions(`team-${teamId}:view-options`);
  const { getTermDisplay } = useTerminology();

  const queryOptions = useMemo(() => {
    const baseOptions = {
      groupBy: viewOptions.groupBy,
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      teamIds: [teamId],
    };

    switch (activeTab) {
      case "all":
        return baseOptions;
      case "active":
        return {
          ...baseOptions,
          categories: ["started" as const],
        };
      case "backlog":
        return {
          ...baseOptions,
          categories: ["backlog" as const],
        };
    }
  }, [
    activeTab,
    viewOptions.groupBy,
    viewOptions.orderBy,
    viewOptions.orderDirection,
    teamId,
  ]);

  const { data: groupedStories, isPending } = useTeamStoriesGrouped(
    teamId!,
    viewOptions.groupBy,
    queryOptions
  );

  if (!viewOptionsLoaded) {
    return (
      <SafeContainer isFull>
        <Header
          viewOptions={viewOptions}
          setViewOptions={setViewOptions}
          resetViewOptions={resetViewOptions}
        />
        <StoriesListSkeleton />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer isFull>
      <Header
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
      />
      <Tabs
        defaultValue={activeTab}
        onValueChange={(value) => setActiveTab(value as TeamStoriesTab)}
      >
        <Tabs.List>
          <Tabs.Tab value="all" className="px-5">
            All {getTermDisplay("storyTerm", { variant: "plural" })}
          </Tabs.Tab>
          <Tabs.Tab value="active" className="px-5">
            Active {getTermDisplay("storyTerm", { variant: "plural" })}
          </Tabs.Tab>
          <Tabs.Tab value="backlog" className="px-5">
            Backlog
          </Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
            visibleColumns={viewOptions.displayColumns}
          />
        </Tabs.Panel>
        <Tabs.Panel value="active">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
            visibleColumns={viewOptions.displayColumns}
          />
        </Tabs.Panel>
        <Tabs.Panel value="backlog">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
            visibleColumns={viewOptions.displayColumns}
          />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
