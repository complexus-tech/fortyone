import React, { useState, useMemo } from "react";
import { Header } from "./components/header";
import {
  SafeContainer,
  Tabs,
  StoriesListSkeleton,
  Text,
} from "@/components/ui";
import { StoriesBoard } from "@/modules/stories/components";
import { useMyStoriesGrouped, useViewOptions } from "./hooks";
import { useTerminology } from "@/hooks/use-terminology";
import type { MyWorkTab } from "./types";

export const MyWork = () => {
  const [activeTab, setActiveTab] = useState<MyWorkTab>("all");
  const { viewOptions, isLoaded: viewOptionsLoaded } = useViewOptions();
  const { getTermDisplay } = useTerminology();

  // Fetch data based on active tab
  const queryOptions = useMemo(() => {
    const baseOptions = {
      groupBy: viewOptions.groupBy,
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
    };

    switch (activeTab) {
      case "all":
        return {
          ...baseOptions,
          assignedToMe: true,
          createdByMe: true,
        };
      case "assigned":
        return {
          ...baseOptions,
          assignedToMe: true,
        };
      case "created":
        return {
          ...baseOptions,
          createdByMe: true,
        };
    }
  }, [
    activeTab,
    viewOptions.groupBy,
    viewOptions.orderBy,
    viewOptions.orderDirection,
  ]);

  const { data: groupedStories, isPending } = useMyStoriesGrouped(
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
        onValueChange={(value) => setActiveTab(value as MyWorkTab)}
      >
        <Tabs.List>
          <Tabs.Tab value="all">
            All {getTermDisplay("storyTerm", { variant: "plural" })}
          </Tabs.Tab>
          <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
          <Tabs.Tab value="created">Created</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <Text>{JSON.stringify(viewOptions)}</Text>
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
          />
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          <StoriesBoard
            groupedStories={groupedStories}
            groupFilters={queryOptions}
            isLoading={isPending}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
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
