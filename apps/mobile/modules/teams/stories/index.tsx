import React, { useState, useMemo } from "react";
import { Header } from "./components";
import { SafeContainer, Tabs, StoriesListSkeleton } from "@/components/ui";
import { StoriesBoard } from "@/modules/stories/components";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks";
import { useTeamViewOptions } from "../hooks";
import { useGlobalSearchParams } from "expo-router";
import type { TeamStoriesTab } from "../types";

export const TeamStories = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const [activeTab, setActiveTab] = useState<TeamStoriesTab>("all");
  const { viewOptions, isLoaded: viewOptionsLoaded } = useTeamViewOptions(
    teamId!
  );

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
        <Header />
        <StoriesListSkeleton />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer isFull>
      <Header />
      <Tabs
        defaultValue={activeTab}
        onValueChange={(value) => setActiveTab(value as TeamStoriesTab)}
      >
        <Tabs.List>
          <Tabs.Tab value="all">All stories</Tabs.Tab>
          <Tabs.Tab value="active">Active</Tabs.Tab>
          <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
          />
        </Tabs.Panel>
        <Tabs.Panel value="active">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
          />
        </Tabs.Panel>
        <Tabs.Panel value="backlog">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
          />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
