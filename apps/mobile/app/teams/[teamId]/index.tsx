import React, { useState, useMemo } from "react";
import { Pressable } from "react-native";
import { useLocalSearchParams } from "expo-router";
import {
  SafeContainer,
  Tabs,
  Text,
  Row,
  Back,
  StoriesListSkeleton,
} from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { StoriesBoard } from "@/modules/stories/components";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks";
import { useTeamViewOptions } from "@/modules/teams/hooks";
import type { TeamStoriesTab } from "@/modules/teams/types";

const StoriesHeader = () => {
  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        Product /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          Stories
        </Text>
      </Text>
      <Pressable
        className="p-2 rounded-md"
        style={({ pressed }) => [
          pressed && { backgroundColor: colors.gray[50] },
        ]}
      >
        <SymbolView name="ellipsis" tintColor={colors.dark[50]} />
      </Pressable>
    </Row>
  );
};

export default function TeamStories() {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  const [activeTab, setActiveTab] = useState<TeamStoriesTab>("all");
  const { viewOptions, isLoaded: viewOptionsLoaded } = useTeamViewOptions(
    teamId!
  );

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
        <StoriesHeader />
        <StoriesListSkeleton />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer isFull>
      <StoriesHeader />
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
}
