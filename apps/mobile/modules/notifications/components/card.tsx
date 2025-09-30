import React from "react";
import { Pressable } from "react-native";
import { Row, Col, Text, Avatar } from "@/components/ui";
import { Ionicons } from "@expo/vector-icons";
import { colors } from "@/constants";
import { StatusDot } from "@/components/story";

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

  const getTypeIcon = () => {
    const message = notification.message.toLowerCase();

    if (notification.type === "story_comment") {
      return (
        <Ionicons
          name="chatbubble-outline"
          size={16}
          color={colors.gray.DEFAULT}
        />
      );
    }

    if (notification.type === "mention") {
      return <Ionicons name="at" size={16} color={colors.gray.DEFAULT} />;
    }

    if (notification.type === "story_update") {
      if (message.includes("deadline") || message.includes("due")) {
        return (
          <Ionicons
            name="calendar-outline"
            size={16}
            color={colors.gray.DEFAULT}
          />
        );
      }
      if (message.includes("status")) {
        return <Ionicons name="ellipse" size={12} color={colors.primary} />;
      }
      if (message.includes("priority")) {
        return (
          <Ionicons name="flag-outline" size={16} color={colors.gray.DEFAULT} />
        );
      }
    }

    return null;
  };

  const formatTimeAgo = (timestamp: string) => {
    // Static time for now - replace with actual formatting later
    return "2m";
  };

  return (
    <Pressable
      style={({ pressed }) => ({
        backgroundColor: pressed ? colors.gray[50] : "white",
        borderBottomWidth: 0.5,
        borderBottomColor: colors.gray[100],
        paddingVertical: 8,
        paddingHorizontal: 16,
      })}
      onPress={onPress}
      onLongPress={onLongPress}
    >
      <Col className="gap-1.5">
        <Row justify="between" align="center">
          <Text
            color={isUnread ? "black" : "muted"}
            className="flex-1 mr-2"
            fontWeight="semibold"
            numberOfLines={1}
          >
            {isUnread && <StatusDot color={colors.primary} size={8} />}{" "}
            {notification.title}
          </Text>
          <Text fontSize="sm" color="muted" className="shrink-0">
            {formatTimeAgo(notification.createdAt)}
          </Text>
        </Row>
        <Row align="center" justify="between" gap={2}>
          <Row align="center" gap={2} className="flex-1">
            <Avatar
              size="sm"
              name={notification.actor.name}
              src={notification.actor.avatar}
              className="shrink-0"
            />
            <Text
              fontSize="sm"
              color={isUnread ? "black" : "muted"}
              className="flex-1"
              numberOfLines={1}
            >
              {notification.message}
            </Text>
          </Row>
          {getTypeIcon()}
        </Row>
      </Col>
    </Pressable>
  );
};
