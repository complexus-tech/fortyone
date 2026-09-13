import React from "react";
import {
  ActivityIndicator,
  View,
  FlatList,
  RefreshControl,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { NotificationCard } from "./card";
import type { AppNotification } from "../types";
import { useTheme } from "@/hooks";
import { colors } from "@/constants/colors";
import { Button, Text } from "@/components/ui";
import { EmptyState } from "./empty-state";

type NotificationListProps = {
  notifications: AppNotification[];
  isLoading?: boolean;
  onRefresh?: () => void;
  hasMore?: boolean;
  isLoadingMore?: boolean;
  onLoadMore?: () => void;
  error?: Error | null;
};

export const NotificationList = ({
  notifications,
  isLoading = false,
  onRefresh,
  hasMore = false,
  isLoadingMore = false,
  onLoadMore,
  error,
}: NotificationListProps) => {
  const insets = useSafeAreaInsets();
  const { resolvedTheme } = useTheme();
  const renderNotification = ({
    item,
    index,
  }: {
    item: AppNotification;
    index: number;
  }) => <NotificationCard {...item} index={index} />;

  return (
    <View className="flex-1">
      <FlatList
        data={notifications}
        renderItem={renderNotification}
        keyExtractor={(item) => item.id}
        ListEmptyComponent={<EmptyState />}
        ListFooterComponent={
          <View className="items-center gap-3 px-4 py-4">
            {error ? <Text color="muted">{error.message}</Text> : null}
            {isLoadingMore ? (
              <ActivityIndicator />
            ) : hasMore ? (
              <Button onPress={onLoadMore}>Load more</Button>
            ) : error ? (
              <Button onPress={onRefresh}>Try again</Button>
            ) : null}
          </View>
        }
        refreshControl={
          <RefreshControl
            refreshing={isLoading}
            onRefresh={onRefresh}
            tintColor={
              resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
            }
            size="default"
          />
        }
        showsVerticalScrollIndicator={false}
        contentContainerStyle={[
          { paddingTop: 0, flexGrow: 1 },
          { paddingBottom: insets.bottom + 20 },
        ]}
        className="flex-1"
      />
    </View>
  );
};
