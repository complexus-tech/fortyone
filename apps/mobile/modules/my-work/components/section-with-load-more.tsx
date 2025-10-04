import React, { useMemo } from "react";
import { View } from "react-native";
import { Story, StorySkeleton } from "@/components/ui";
import { StoryGroupHeader } from "./story-group-header";
import { SectionFooter } from "./section-footer";
import { useGroupStoriesInfinite } from "@/modules/stories/hooks";
import type { MyWorkSection } from "../types";
import type { GroupStoryParams } from "@/modules/stories/types";

type SectionWithLoadMoreProps = {
  section: MyWorkSection;
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
    <View>
      <StoryGroupHeader title={section.title} color={section.color} />
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
