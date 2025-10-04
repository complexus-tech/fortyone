import React, { useState, useMemo } from "react";
import { Header } from "./components/header";
import { GroupedStoriesList } from "./components/grouped-list";
import { SafeContainer, Tabs, StoriesListSkeleton } from "@/components/ui";
import { useMyStoriesGrouped, useViewOptions } from "./hooks";
import { useStatuses } from "@/modules/statuses";
import { useMembers } from "@/modules/members";
import type { MyWorkTab, MyWorkSection } from "./types";

export const MyWork = () => {
  const [activeTab, setActiveTab] = useState<MyWorkTab>("all");
  const { viewOptions, isLoaded: viewOptionsLoaded } = useViewOptions();

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

    return groupedStories.groups
      .map((group) => {
        let title = "";
        let color: string | undefined;

        switch (viewOptions.groupBy) {
          case "status": {
            const status = statuses.find((s) => s.id === group.key);
            title = status?.name || `Unknown (${group.key})`;
            color = status?.color;
            break;
          }
          case "priority": {
            title = group.key;
            break;
          }
          case "assignee": {
            const member = members.find((m) => m.id === group.key);
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
      .filter((section) => section.data.length > 0);
  }, [groupedStories, viewOptions.groupBy, statuses, members]);

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
          <Tabs.Tab value="all">All stories</Tabs.Tab>
          <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
          <Tabs.Tab value="created">Created</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <GroupedStoriesList
            sections={sections}
            groupFilters={queryOptions}
            isLoading={isPending}
          />
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          <GroupedStoriesList
            sections={sections}
            groupFilters={queryOptions}
            isLoading={isPending}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <GroupedStoriesList
            sections={sections}
            groupFilters={queryOptions}
            isLoading={isPending}
          />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
