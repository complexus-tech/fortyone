import React, { useState, useMemo } from "react";
import { useRouter } from "expo-router";
import { Header } from "./components/header";
import { GroupedStoriesList } from "./components/grouped-list";
import { SafeContainer, Tabs } from "@/components/ui";
import { useMyStoriesGrouped, useViewOptions } from "./hooks";
import { useStatuses } from "@/modules/statuses";
import { useMembers } from "@/modules/members";
import type { MyWorkTab, MyWorkSection } from "./types";
import type { StoryPriority } from "@/modules/stories/types";

const PRIORITY_ORDER: Record<StoryPriority, number> = {
  "No Priority": 0,
  Low: 1,
  Medium: 2,
  High: 3,
  Urgent: 4,
};

const PRIORITY_LABELS: Record<StoryPriority, string> = {
  "No Priority": "No Priority",
  Low: "Low",
  Medium: "Medium",
  High: "High",
  Urgent: "Urgent",
};

export const MyWork = () => {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<MyWorkTab>("all");
  const { viewOptions, isLoaded: viewOptionsLoaded } = useViewOptions();

  // Fetch data based on active tab
  const queryOptions = useMemo(() => {
    const baseOptions = {
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
  }, [activeTab, viewOptions.orderBy, viewOptions.orderDirection]);

  const { data: groupedStories, isPending: isStoriesPending } =
    useMyStoriesGrouped(viewOptions.groupBy, queryOptions);

  const { data: statuses = [], isPending: isStatusesPending } = useStatuses();
  const { data: members = [], isPending: isMembersPending } = useMembers();

  // Determine if we're still loading based on what we're grouping by
  const isPending = useMemo(() => {
    if (isStoriesPending) return true;
    if (viewOptions.groupBy === "status" && isStatusesPending) return true;
    if (viewOptions.groupBy === "assignee" && isMembersPending) return true;
    return false;
  }, [
    isStoriesPending,
    isStatusesPending,
    isMembersPending,
    viewOptions.groupBy,
  ]);

  // Transform grouped stories into sections for SectionList
  const sections = useMemo<MyWorkSection[]>(() => {
    if (!groupedStories?.groups) return [];

    console.log("🔍 Transforming sections:", {
      groupBy: viewOptions.groupBy,
      groupsCount: groupedStories.groups.length,
      statusesCount: statuses.length,
      membersCount: members.length,
    });

    return groupedStories.groups
      .map((group) => {
        let title = "";
        let color: string | undefined;

        // Determine title and color based on groupBy type
        switch (viewOptions.groupBy) {
          case "status": {
            const status = statuses.find((s) => s.id === group.key);
            console.log(`📊 Status lookup for key "${group.key}":`, {
              found: !!status,
              statusName: status?.name,
              allStatusIds: statuses.map((s) => s.id),
            });
            title = status?.name || `Unknown (${group.key})`;
            color = status?.color;
            break;
          }
          case "priority": {
            title = PRIORITY_LABELS[group.key as StoryPriority] || group.key;
            break;
          }
          case "assignee": {
            const member = members.find((m) => m.id === group.key);
            console.log(`👤 Member lookup for key "${group.key}":`, {
              found: !!member,
              memberName: member?.fullName || member?.username,
            });
            title = member?.fullName || member?.username || "Unassigned";
            break;
          }
        }

        return {
          title,
          color,
          data: group.stories,
          key: group.key,
          totalCount: group.totalCount,
          loadedCount: group.loadedCount,
          hasMore: group.hasMore,
        };
      })
      .filter((section) => section.data.length > 0); // Only show sections with items on mobile
  }, [groupedStories, viewOptions.groupBy, statuses, members]);

  const handleStoryPress = (storyId: string) => {
    router.push(`/story/${storyId}`);
  };

  // Wait for view options to load from storage
  if (!viewOptionsLoaded) {
    return <SafeContainer isFull>{/* Loading skeleton */}</SafeContainer>;
  }

  return (
    <SafeContainer isFull>
      <Header />
      <Tabs
        defaultValue={activeTab}
        onValueChange={(value) => setActiveTab(value as MyWorkTab)}
      >
        <Tabs.List>
          <Tabs.Tab value="all">All stories</Tabs.Tab>
          <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
          <Tabs.Tab value="created">Created</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <GroupedStoriesList
            sections={sections}
            onStoryPress={handleStoryPress}
            isLoading={isPending}
          />
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          <GroupedStoriesList
            sections={sections}
            onStoryPress={handleStoryPress}
            isLoading={isPending}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <GroupedStoriesList
            sections={sections}
            onStoryPress={handleStoryPress}
            isLoading={isPending}
          />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
