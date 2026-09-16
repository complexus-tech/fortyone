import type { ReactNode } from "react";
import { FlatList, View, ActivityIndicator, Linking } from "react-native";
import Markdown from "react-native-markdown-display";
import { formatDistanceToNow } from "date-fns";
import { toast } from "sonner-native";
import { Avatar, Button, Row, Tabs, Text } from "@/components/ui";
import { RichTextViewer } from "@/components/rich-text/viewer";
import { isSafeLink, plainTextToHtml } from "@/components/rich-text/content";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

export type DetailFeedRow = {
  id: string;
  kind: "comment" | "update";
  author: string;
  avatar?: string | null;
  body: string;
  format?: "html" | "markdown" | "plain";
  createdAt: string;
  deliveryStatus?: string;
  parentId?: string | null;
  onReply?: () => void;
};
export function MarkdownContent({ children }: { children: string }) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  return (
    <Markdown
      style={{
        body: {
          color: theme.foreground,
          fontSize: 17,
          lineHeight: 25,
          fontWeight: "500",
        },
        link: { color: theme.foreground, textDecorationLine: "underline" },
      }}
      onLinkPress={(url) => {
        if (isSafeLink(url))
          void Linking.openURL(url).catch(() =>
            toast.error("Could not open this link."),
          );
        return false;
      }}
    >
      {children}
    </Markdown>
  );
}
function FeedRow({ item }: { item: DetailFeedRow }) {
  const timestamp = Number.isNaN(Date.parse(item.createdAt))
    ? ""
    : formatDistanceToNow(new Date(item.createdAt), { addSuffix: true });
  if (item.kind === "update")
    return (
      <View className="flex-row items-start gap-[10px] py-[10px]">
        <View
          accessible={false}
          className="mt-[8px] h-[8px] w-[8px] rounded-full border border-gray-300 dark:border-gray-400"
        />
        <View className="min-w-0 flex-1 gap-[4px]">
          <Text color="muted" style={{ fontSize: 16, lineHeight: 24 }}>
            <Text fontWeight="semibold">{item.author}</Text>
            {` ${item.body}`}
          </Text>
          <Text color="muted" fontSize="xs">
            {timestamp}
          </Text>
        </View>
      </View>
    );
  return (
    <View
      className="my-[6px] rounded-[20px] bg-gray-50 p-[14px] dark:bg-dark-100"
      style={item.parentId ? { marginLeft: 20 } : undefined}
    >
      <Row align="center" wrap>
        <Avatar
          name={item.author}
          src={item.avatar}
          size="xs"
          className="mr-2"
        />
        <Text fontSize="sm" fontWeight="bold">
          {item.author}
        </Text>
        <Text fontSize="xs" color="muted" className="ml-2">
          {timestamp}
        </Text>
      </Row>
      <View className="mt-[6px]">
        {item.format === "markdown" ? (
          <MarkdownContent>{item.body}</MarkdownContent>
        ) : (
          <RichTextViewer
            html={
              item.format === "html" ? item.body : plainTextToHtml(item.body)
            }
          />
        )}
      </View>
      {item.deliveryStatus ? (
        <Text
          fontSize="xs"
          color={
            item.deliveryStatus === "failed" ||
            item.deliveryStatus === "not-sent"
              ? "danger"
              : "muted"
          }
        >
          {item.deliveryStatus}
        </Text>
      ) : null}
      {item.onReply ? (
        <Button color="tertiary" fullWidth={false} onPress={item.onReply}>
          Reply
        </Button>
      ) : null}
    </View>
  );
}
export function DetailFeed({
  children,
  rows,
  tab,
  onTabChange,
  options,
  loading,
  error,
  onRetry,
  hasMore,
  loadingMore,
  onLoadMore,
  refreshing,
  onRefresh,
}: {
  children: ReactNode;
  rows: DetailFeedRow[];
  tab: string;
  onTabChange?: (value: string) => void;
  options: { value: string; label: string }[];
  loading?: boolean;
  error?: Error | null;
  onRetry?: () => void;
  hasMore?: boolean;
  loadingMore?: boolean;
  onLoadMore?: () => void;
  refreshing?: boolean;
  onRefresh?: () => void;
}) {
  return (
    <Tabs value={tab} onValueChange={onTabChange}>
      <FlatList
        className="flex-1"
        data={rows}
        keyExtractor={(row) => row.id}
        renderItem={({ item }) => (
          <View className="px-[20px]">
            <FeedRow item={item} />
          </View>
        )}
        contentContainerStyle={{ paddingBottom: 24 }}
        keyboardShouldPersistTaps="handled"
        initialNumToRender={4}
        maxToRenderPerBatch={4}
        windowSize={3}
        removeClippedSubviews={false}
        refreshing={refreshing ?? false}
        onRefresh={onRefresh}
        ListHeaderComponent={
          <>
            {children}
            {options.length ? (
              <Tabs.List
                className="mb-[6px] pt-[4px]"
                accessibilityLabel="Activity"
                labelSize={14}
                options={options}
              />
            ) : null}
          </>
        }
        ListFooterComponent={
          options.length ? (
            <View className="px-[20px] gap-3 py-4">
              {loading ? (
                <ActivityIndicator accessibilityLabel="Loading activity" />
              ) : error ? (
                <>
                  <Text color="danger">{error.message}</Text>
                  <Button color="tertiary" onPress={onRetry}>
                    Try again
                  </Button>
                </>
              ) : (
                <>
                  {!rows.length ? (
                    <Text color="muted" fontSize="sm">
                      {tab === "updates" ? "No updates yet" : "No comments yet"}
                    </Text>
                  ) : null}
                  {hasMore ? (
                    <Button
                      color="tertiary"
                      loading={loadingMore}
                      onPress={onLoadMore}
                    >
                      Load more
                    </Button>
                  ) : null}
                </>
              )}
            </View>
          ) : null
        }
      />
    </Tabs>
  );
}
