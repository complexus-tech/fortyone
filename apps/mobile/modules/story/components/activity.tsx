import type { DetailedStory, StoryActivity } from "@/modules/stories/types";
import type { Comment } from "@/types";
import type { ListRenderItemInfo } from "react-native";
import React, { useMemo, useState } from "react";
import { ActivityIndicator, FlatList, View } from "react-native";
import { Button, Tabs, Text } from "@/components/ui";
import { useStoryActivitiesInfinite } from "../hooks/use-story-activities";
import { useStoryCommentsInfinite } from "../hooks/use-story-comments";
import { ActivityItem } from "./activity-item";
import { CommentItem } from "./comment-item";
import { CommentComposer } from "./comment-composer";

type FeedTab = "updates" | "comments";
type FeedRow =
  | { type: "update"; activity: StoryActivity }
  | { type: "comment"; comment: Comment };

const rowKey = (row: FeedRow) =>
  row.type === "comment"
    ? `comment:${row.comment.id}`
    : `update:${row.activity.id}`;

const FeedFooter = ({
  tab,
  loading,
  error,
  empty,
  hasMore,
  fetchingMore,
  retrying,
  onRetry,
  onLoadMore,
}: {
  tab: FeedTab;
  loading: boolean;
  error: Error | null;
  empty: boolean;
  hasMore: boolean;
  fetchingMore: boolean;
  retrying: boolean;
  onRetry: () => void;
  onLoadMore: () => void;
}) => (
  <View className="px-4.5 gap-3 py-3">
    {loading ? (
      <ActivityIndicator accessibilityLabel={`Loading ${tab}`} />
    ) : error ? (
      <>
        <Text color="danger" accessibilityRole="alert">
          {error.message || `Could not load ${tab}.`}
        </Text>
        <Button color="tertiary" loading={retrying} onPress={onRetry}>
          Try again
        </Button>
      </>
    ) : (
      <>
        {empty ? (
          <Text>
            {tab === "comments" ? "No comments yet" : "No updates available"}
          </Text>
        ) : null}
        {hasMore ? (
          <Button color="tertiary" loading={fetchingMore} onPress={onLoadMore}>
            {`Load more ${tab}`}
          </Button>
        ) : null}
      </>
    )}
  </View>
);

// One vertical list owns the detail header and both feeds. In particular,
// off-window comments unmount their DOM viewers instead of accumulating WebViews.
export const Activity = ({
  story,
  children,
}: {
  story: DetailedStory;
  children: React.ReactNode;
}) => {
  const [activeTab, setActiveTab] = useState<FeedTab>("updates");
  const activities = useStoryActivitiesInfinite(
    story.id,
    activeTab === "updates",
  );
  const comments = useStoryCommentsInfinite(story.id, activeTab === "comments");
  const feed = activeTab === "comments" ? comments : activities;
  const rows = useMemo<FeedRow[]>(() => {
    if (activeTab === "comments") {
      return (
        comments.data?.pages.flatMap((page) =>
          page.comments.map((comment) => ({
            type: "comment" as const,
            comment,
          })),
        ) ?? []
      );
    }
    return (
      activities.data?.pages.flatMap((page) =>
        page.activities
          .filter((activity) => activity.field !== "completed_at")
          .map((activity) => ({ type: "update" as const, activity })),
      ) ?? []
    );
  }, [activeTab, activities.data, comments.data]);

  const renderItem = ({ item, index }: ListRenderItemInfo<FeedRow>) => (
    <View className="px-4.5">
      {item.type === "comment" ? (
        <CommentItem {...item.comment} />
      ) : (
        <ActivityItem
          {...item.activity}
          teamId={story.teamId}
          isTimeShown={index === 0 || index === rows.length - 1}
        />
      )}
    </View>
  );

  const loadMore = () => {
    if (feed.hasNextPage && !feed.isFetching) void feed.fetchNextPage();
  };
  const retry = () => {
    if (feed.isFetching) return;
    if (feed.isFetchNextPageError) void feed.fetchNextPage();
    else void feed.refetch();
  };

  return (
    <Tabs
      value={activeTab}
      onValueChange={(value) => {
        if (value === "updates" || value === "comments") setActiveTab(value);
      }}
    >
      <FlatList
        className="flex-1"
        data={rows}
        keyExtractor={rowKey}
        renderItem={renderItem}
        contentContainerStyle={{ paddingBottom: 100 }}
        keyboardShouldPersistTaps="handled"
        initialNumToRender={4}
        maxToRenderPerBatch={4}
        windowSize={3}
        // Native clipping can blank WebViews when they are reattached. React's
        // list window still unmounts distant rows with clipping disabled.
        removeClippedSubviews={false}
        ListHeaderComponent={
          <>
            {children}
            <Tabs.List className="mb-2">
              <Tabs.Tab value="updates">Updates</Tabs.Tab>
              <Tabs.Tab value="comments">Comments</Tabs.Tab>
            </Tabs.List>
            {activeTab === "comments" ? (
              <View className="px-4.5">
                <CommentComposer storyId={story.id} />
              </View>
            ) : null}
          </>
        }
        ListFooterComponent={
          <FeedFooter
            tab={activeTab}
            loading={feed.isPending}
            error={feed.error}
            empty={rows.length === 0}
            hasMore={Boolean(feed.hasNextPage)}
            fetchingMore={feed.isFetchingNextPage}
            retrying={feed.isFetching}
            onRetry={retry}
            onLoadMore={loadMore}
          />
        }
      />
    </Tabs>
  );
};
