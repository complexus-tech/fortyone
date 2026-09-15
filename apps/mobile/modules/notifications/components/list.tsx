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
import { themeColors } from "@/constants/colors";
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

const renderNotification = ({ item }: { item: AppNotification }) => (
  <NotificationCard {...item} />
);
const notificationKey = (item: AppNotification) => item.id;

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

  return (
    <View className="flex-1">
      <FlatList
        data={notifications}
        renderItem={renderNotification}
        keyExtractor={notificationKey}
        keyboardShouldPersistTaps="handled"
        initialNumToRender={12}
        maxToRenderPerBatch={10}
        ListEmptyComponent={<EmptyState />}
        ListFooterComponent={
          <View className="items-center gap-3 px-[20px] py-4">
            {error ? (
              <Text color="muted" fontSize="sm" align="center">
                {error.message}
              </Text>
            ) : null}
            {isLoadingMore ? (
              <ActivityIndicator />
            ) : hasMore ? (
              <Button color="tertiary" onPress={onLoadMore}>
                Load more
              </Button>
            ) : error ? (
              <Button color="tertiary" onPress={onRefresh}>
                Try again
              </Button>
            ) : null}
          </View>
        }
        refreshControl={
          <RefreshControl
            refreshing={isLoading}
            onRefresh={onRefresh}
            tintColor={themeColors[resolvedTheme].textMuted}
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
