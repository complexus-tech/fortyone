import React, { useMemo } from "react";
import { ScrollView } from "react-native";
import { StoriesListSkeleton } from "@/components/ui";
import { SectionWithLoadMore, EmptyState } from "./";
import { useStatuses } from "@/modules/statuses";
import { useMembers } from "@/modules/members";
import type { GroupedStoriesResponse, GroupStoryParams, Story } from "../types";

type StoriesSection = {
  title: string;
  color?: string;
  data: Story[];
  key: string;
  totalCount: number;
  loadedCount: number;
  hasMore: boolean;
};

type StoriesBoardProps = {
  groupedStories?: GroupedStoriesResponse;
  groupFilters: Omit<GroupStoryParams, "groupKey">;
  isLoading?: boolean;
  emptyTitle?: string;
  emptyMessage?: string;
};

export const StoriesBoard = ({
  groupedStories,
  groupFilters,
  isLoading = false,
  emptyTitle = "No stories found",
  emptyMessage = "There are no stories to display at the moment.",
}: StoriesBoardProps) => {
  const { data: statuses = [], isPending: isStatusesPending } = useStatuses();
  const { data: members = [], isPending: isMembersPending } = useMembers();

  const groupBy = groupFilters.groupBy;

  const isPending = useMemo(() => {
    if (isLoading) return true;
    if (groupBy === "status" && isStatusesPending) return true;
    if (groupBy === "assignee" && isMembersPending) return true;
    return false;
  }, [isLoading, groupBy, isStatusesPending, isMembersPending]);

  const sections = useMemo<StoriesSection[]>(() => {
    if (!groupedStories?.groups) return [];

    return groupedStories.groups
      .map((group) => {
        let title = "";
        let color: string | undefined;

        switch (groupBy) {
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
  }, [groupedStories, groupBy, statuses, members]);

  if (isPending) {
    return <StoriesListSkeleton />;
  }

  if (sections.length === 0) {
    return <EmptyState title={emptyTitle} message={emptyMessage} />;
  }

  return (
    <ScrollView
      showsVerticalScrollIndicator={false}
      contentContainerStyle={{ paddingBottom: 24 }}
    >
      {sections.map((section) => (
        <SectionWithLoadMore
          key={section.key}
          section={section}
          groupFilters={groupFilters}
        />
      ))}
    </ScrollView>
  );
};
