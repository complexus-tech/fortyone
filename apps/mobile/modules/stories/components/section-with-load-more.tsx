import React, { useMemo } from "react";
import { FlatList, View } from "react-native";
import { Story, StorySkeleton, Text } from "@/components/ui";
import { SectionFooter } from "./section-footer";
import { useGroupStoriesInfinite } from "../hooks";
import type { GroupStoryParams, Story as StoryType } from "../types";
import { DisplayColumn } from "@/types/stories-view-options";

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
  visibleColumns: DisplayColumn[];
};

export const SectionWithLoadMore = ({
  section,
  groupFilters,
  visibleColumns,
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

  const renderItem = ({ item }: { item: StoryType }) => (
    <Story story={item} visibleColumns={visibleColumns} />
  );

  const renderHeader = () => (
    <Text
      className="px-4 pt-2 pb-1"
      fontSize="sm"
      fontWeight="semibold"
      color="muted"
    >
      {section.title}
    </Text>
  );

  const renderFooter = () => (
    <>
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
    </>
  );

  return (
    <View className="mb-2">
      <FlatList
        data={allStories}
        renderItem={renderItem}
        keyExtractor={(item) => item.id}
        ListHeaderComponent={renderHeader}
        ListFooterComponent={renderFooter}
        showsVerticalScrollIndicator={false}
        scrollEnabled={false}
        initialNumToRender={10}
        maxToRenderPerBatch={5}
        windowSize={10}
        removeClippedSubviews={true}
      />
    </View>
  );
};
