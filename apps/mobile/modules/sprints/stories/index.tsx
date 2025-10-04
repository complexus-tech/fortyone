import React, { useState, useMemo } from "react";
import { Header } from "./components";
import { SafeContainer, Tabs, StoriesListSkeleton } from "@/components/ui";
import { StoriesBoard } from "@/modules/stories/components";
import { useSprintStoriesGrouped } from "@/modules/stories/hooks";
import { useSprintViewOptions } from "./hooks";
import { useGlobalSearchParams } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";
import type { SprintStoriesTab } from "./types";

export const SprintStories = () => {
  const { sprintId, teamId } = useGlobalSearchParams<{
    sprintId: string;
    teamId: string;
  }>();
  const [activeTab, setActiveTab] = useState<SprintStoriesTab>("all");
  const { viewOptions, isLoaded: viewOptionsLoaded } = useSprintViewOptions(
    sprintId!
  );
  const { getTermDisplay } = useTerminology();

  const queryOptions = useMemo(() => {
    const baseOptions = {
      groupBy: viewOptions.groupBy,
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      teamIds: [teamId!],
    };

    switch (activeTab) {
      case "all":
        return baseOptions;
      case "active":
        return {
          ...baseOptions,
          categories: ["started" as const, "paused" as const],
        };
      case "completed":
        return {
          ...baseOptions,
          categories: ["completed" as const],
        };
    }
  }, [
    activeTab,
    viewOptions.groupBy,
    viewOptions.orderBy,
    viewOptions.orderDirection,
    teamId,
  ]);

  const { data: groupedStories, isPending } = useSprintStoriesGrouped(
    sprintId!,
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
        onValueChange={(value) => setActiveTab(value as SprintStoriesTab)}
      >
        <Tabs.List>
          <Tabs.Tab value="all">
            All {getTermDisplay("storyTerm", { variant: "plural" })}
          </Tabs.Tab>
          <Tabs.Tab value="active">Active</Tabs.Tab>
          <Tabs.Tab value="completed">Completed</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
            emptyTitle={`No ${getTermDisplay("storyTerm", { variant: "plural" })} found in this sprint`}
            emptyMessage={`There are no ${getTermDisplay("storyTerm", { variant: "plural" })} in this sprint at the moment.`}
          />
        </Tabs.Panel>
        <Tabs.Panel value="active">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
            emptyTitle={`No active ${getTermDisplay("storyTerm", { variant: "plural" })} found`}
            emptyMessage={`There are no active ${getTermDisplay("storyTerm", { variant: "plural" })} in this sprint.`}
          />
        </Tabs.Panel>
        <Tabs.Panel value="completed">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
            emptyTitle={`No completed ${getTermDisplay("storyTerm", { variant: "plural" })} found`}
            emptyMessage={`There are no completed ${getTermDisplay("storyTerm", { variant: "plural" })} in this sprint.`}
          />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
