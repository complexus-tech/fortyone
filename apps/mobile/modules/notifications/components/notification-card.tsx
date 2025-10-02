import React from "react";
import { Pressable } from "react-native";
import { Row, Col, Text, Avatar } from "@/components/ui";
import { Ionicons } from "@expo/vector-icons";
import { colors } from "@/constants";
import type { AppNotification } from "../types";
import { useMembers } from "@/modules/members";
import { renderTemplate } from "../utils/render-template";
import { useRouter } from "expo-router";

type NotificationCardProps = AppNotification & {
  index: number;
};

export const NotificationCard = ({
  id,
  title,
  message,
  type,
  entityId,
  entityType,
  readAt,
  createdAt,
  actorId,
  index,
}: NotificationCardProps) => {
  const router = useRouter();
  const { data: members = [] } = useMembers();
  const actor = members.find((member) => member.id === actorId);
  const isUnread = !readAt;

  const handlePress = () => {
    router.push(`/story/${entityId}`);
  };

  const { text } = renderTemplate(message);

  const getTypeIcon = () => {
    const messageText = text.toLowerCase();

    if (type === "story_comment") {
      return (
        <Ionicons
          name="chatbubble-outline"
          size={16}
          color={colors.gray.DEFAULT}
        />
      );
    }

    if (type === "mention") {
      return <Ionicons name="at" size={16} color={colors.gray.DEFAULT} />;
    }

    if (type === "story_update") {
      if (messageText.includes("deadline") || messageText.includes("due")) {
        return (
          <Ionicons
            name="calendar-outline"
            size={16}
            color={colors.gray.DEFAULT}
          />
        );
      }
      if (messageText.includes("status")) {
        return <Ionicons name="ellipse" size={12} color={colors.primary} />;
      }
      if (messageText.includes("priority")) {
        return (
          <Ionicons name="flag-outline" size={16} color={colors.gray.DEFAULT} />
        );
      }
    }

    return null;
  };

  const formatTimeAgo = (timestamp: string) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diffInMinutes = Math.floor(
      (now.getTime() - time.getTime()) / (1000 * 60)
    );

    if (diffInMinutes < 1) return "now";
    if (diffInMinutes < 60) return `${diffInMinutes}m`;
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}h`;
    return `${Math.floor(diffInMinutes / 1440)}d`;
  };

  return (
    <Pressable
      style={({ pressed }) => ({
        backgroundColor: pressed ? colors.gray[50] : "white",
        paddingVertical: 12,
        paddingHorizontal: 16,
      })}
      onPress={handlePress}
    >
      <Row align="center" gap={2}>
        <Avatar
          size="lg"
          name={actor?.fullName || actor?.username || "Unknown"}
          src={actor?.avatarUrl}
          className="shrink-0"
        />
        <Col flex={1} gap={1}>
          <Row justify="between" align="center">
            <Text
              color={isUnread ? "black" : "muted"}
              className="flex-1 mr-2"
              fontWeight={isUnread ? "bold" : "medium"}
              numberOfLines={1}
            >
              {title}
            </Text>
            <Text
              fontSize="sm"
              color={isUnread ? "black" : "muted"}
              className="shrink-0"
            >
              {formatTimeAgo(createdAt)}
            </Text>
          </Row>
          <Row align="center" justify="between" gap={2}>
            <Row align="center" gap={2} className="flex-1">
              <Text
                fontSize="sm"
                color={isUnread ? "black" : "muted"}
                className="flex-1"
                fontWeight={isUnread ? "semibold" : "medium"}
                numberOfLines={1}
              >
                {text}
              </Text>
            </Row>
            {getTypeIcon()}
          </Row>
        </Col>
      </Row>
    </Pressable>
  );
};
