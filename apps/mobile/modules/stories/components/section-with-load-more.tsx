import React, { useMemo } from "react";
import { View } from "react-native";
import { Story, StorySkeleton, Text } from "@/components/ui";
import { SectionFooter } from "./section-footer";
import { useGroupStoriesInfinite } from "../hooks";
import type { GroupStoryParams, Story as StoryType } from "../types";

type StoriesSection = {
  title: string;
  color?: string;
  data: StoryType[];
  key: string;
  totalCount: number;
  loadedCount: number;
  hasMore: boolean;
};

type SectionWithLoadMoreProps = {
  section: StoriesSection;
  groupFilters: Omit<GroupStoryParams, "groupKey">;
};

export const SectionWithLoadMore = ({
  section,
  groupFilters,
}: SectionWithLoadMoreProps) => {
  const params: GroupStoryParams = {
    groupKey: section.key,
    ...groupFilters,
  };

  const initialGroup = {
    key: section.key,
    stories: section.data,
    loadedCount: section.loadedCount,
    totalCount: section.totalCount,
    hasMore: section.hasMore,
    nextPage: 2,
  };

  const {
    data: infiniteData,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useGroupStoriesInfinite(params, initialGroup);

  const allStories = useMemo(
    () =>
      infiniteData?.pages.flatMap((page) => page?.stories ?? []) ||
      section.data,
    [infiniteData, section.data]
  );

  const loadedCount = allStories.length;

  return (
    <View className="mb-2">
      <Text className="px-4 pt-2 pb-1" fontWeight="semibold" color="muted">
        {section.title}
      </Text>
      {allStories.map((story) => (
        <Story key={story.id} {...story} />
      ))}
      {isFetchingNextPage && (
        <>
          <StorySkeleton />
          <StorySkeleton />
          <StorySkeleton />
        </>
      )}
      <SectionFooter
        hasMore={hasNextPage || false}
        loadedCount={loadedCount}
        totalCount={section.totalCount}
        isLoading={isFetchingNextPage}
        onLoadMore={() => fetchNextPage()}
      />
    </View>
  );
};
