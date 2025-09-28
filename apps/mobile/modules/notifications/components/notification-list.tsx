import React from "react";
import { View, FlatList, RefreshControl } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { NotificationCard } from "./notification-card";

type Notification = {
  id: string;
  title: string;
  message: string;
  type: "story_update" | "story_comment" | "mention";
  actor: {
    name: string;
    avatar?: string;
  };
  createdAt: string;
  readAt: string | null;
  entityId: string;
};

type NotificationListProps = {
  notifications: Notification[];
  isLoading?: boolean;
  onRefresh?: () => void;
  onNotificationPress?: (notification: Notification) => void;
  onNotificationLongPress?: (notification: Notification) => void;
};

export const NotificationList = ({
  notifications,
  isLoading = false,
  onRefresh,
  onNotificationPress,
  onNotificationLongPress,
}: NotificationListProps) => {
  const insets = useSafeAreaInsets();

  const renderNotification = ({
    item,
    index,
  }: {
    item: Notification;
    index: number;
  }) => (
    <NotificationCard
      notification={item}
      index={index}
      onPress={() => onNotificationPress?.(item)}
      onLongPress={() => onNotificationLongPress?.(item)}
    />
  );

  return (
    <View className="flex-1 bg-gray-50">
      <FlatList
        data={notifications}
        renderItem={renderNotification}
        keyExtractor={(item) => item.id}
        refreshControl={
          <RefreshControl
            refreshing={isLoading}
            onRefresh={onRefresh}
            tintColor="#007AFF"
            colors={["#007AFF"]}
          />
        }
        showsVerticalScrollIndicator={false}
        contentContainerStyle={[
          { paddingTop: 0 },
          { paddingBottom: insets.bottom + 20 },
        ]}
        className="flex-1"
      />
    </View>
  );
};
