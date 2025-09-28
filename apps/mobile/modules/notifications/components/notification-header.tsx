import React from "react";
import { View, Pressable } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Text } from "@/components/ui";

type NotificationHeaderProps = {
  unreadCount?: number;
  onFilterPress?: () => void;
  onMarkAllRead?: () => void;
  onDeleteAll?: () => void;
};

export const NotificationHeader = ({
  unreadCount = 0,
  onFilterPress,
  onMarkAllRead,
  onDeleteAll,
}: NotificationHeaderProps) => {
  const insets = useSafeAreaInsets();

  return (
    <View className="bg-transparent" style={{ paddingTop: insets.top }}>
      <View className="bg-gray-50 border-b border-gray-200">
        <View className="flex-row justify-between items-start px-4 pt-2 pb-2">
          <View className="flex-row items-center">
            <Text fontSize="2xl" fontWeight="semibold" color="black">
              Notifications
            </Text>
          </View>

          <View className="flex-row items-center">
            <Pressable
              className="px-2 py-2 rounded-md"
              style={({ pressed }) => [
                pressed && { backgroundColor: "#F2F2F7" },
              ]}
              onPress={onFilterPress}
            >
              <Text fontSize="lg" fontWeight="semibold" color="black">
                ⋯
              </Text>
            </Pressable>
          </View>
        </View>
      </View>
    </View>
  );
};
