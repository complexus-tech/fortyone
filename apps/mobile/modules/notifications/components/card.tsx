import React from "react";
import { View, Pressable } from "react-native";
import { Text } from "@/components/ui";

type NotificationCardProps = {
  notification: {
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
    repository?: string;
    issueNumber?: number;
    status?: string;
  };
  index: number;
  onPress?: () => void;
  onLongPress?: () => void;
};

export const NotificationCard = ({
  notification,
  onPress,
  onLongPress,
}: NotificationCardProps) => {
  const isUnread = !notification.readAt;

  const getStatusIcon = () => {
    if (notification.status === "closed") return "✕";
    if (notification.status === "merged") return "✓";
    if (notification.status === "urgent") return "!";
    return "→";
  };

  const getStatusColor = () => {
    if (notification.status === "closed") return "#6D6D70";
    if (notification.status === "urgent") return "#FF9500";
    return "#6D6D70";
  };

  return (
    <Pressable
      className={`px-4 py-3 border-b border-gray-100 ${
        isUnread ? "bg-gray-50" : "bg-white"
      }`}
      style={({ pressed }) => [pressed && { backgroundColor: "#F2F2F7" }]}
      onPress={onPress}
      onLongPress={onLongPress}
    >
      <View className="flex-row">
        <View className="mr-3 items-center">
          <View className="items-center justify-center relative">
            <View className="w-10 h-10 rounded-full bg-black justify-center items-center">
              <Text fontSize="md" fontWeight="semibold" color="white">
                {notification.actor.name.charAt(0).toUpperCase()}
              </Text>
            </View>
            <View
              className="absolute -bottom-0.5 -right-0.5 w-4 h-4 rounded-full justify-center items-center"
              style={{ backgroundColor: getStatusColor() }}
            >
              <Text fontSize="xs" fontWeight="semibold" color="white">
                {getStatusIcon()}
              </Text>
            </View>
          </View>
        </View>

        <View className="flex-1">
          <View className="flex-row justify-between items-center mb-0.5">
            <Text
              fontSize="md"
              fontWeight="semibold"
              color="black"
              className="flex-1 mr-2"
              numberOfLines={1}
            >
              {notification.title}
            </Text>
            <Text fontSize="sm" color="muted">
              2mo ago
            </Text>
          </View>

          {notification.message && (
            <Text
              fontSize="sm"
              color="muted"
              className="leading-5"
              numberOfLines={1}
            >
              {notification.message}
            </Text>
          )}
        </View>
      </View>
    </Pressable>
  );
};
