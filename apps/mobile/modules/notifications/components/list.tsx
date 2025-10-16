import React from "react";
import { View, FlatList, RefreshControl } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { NotificationCard } from "./card";
import type { AppNotification } from "../types";
import { useTheme } from "@/hooks";
import { colors } from "@/constants/colors";

type NotificationListProps = {
  notifications: AppNotification[];
  isLoading?: boolean;
  onRefresh?: () => void;
};

export const NotificationList = ({
  notifications,
  isLoading = false,
  onRefresh,
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
        refreshControl={
          <RefreshControl
            refreshing={isLoading}
            onRefresh={onRefresh}
            tintColor={
              resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
            }
            size={20}
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
