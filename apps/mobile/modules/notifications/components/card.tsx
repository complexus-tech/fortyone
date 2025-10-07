import React from "react";
import { Pressable } from "react-native";
import { Row, Col, Text, Avatar } from "@/components/ui";
import { colors } from "@/constants";
import type { AppNotification } from "../types";
import { useMembers } from "@/modules/members";
import { renderTemplateJSX } from "../utils/render-template";
import { useRouter } from "expo-router";
import { Dot } from "@/components/icons";
import { useTerminology } from "@/hooks";
import { useReadNotificationMutation } from "../hooks";

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
  const { getTermDisplay } = useTerminology();
  const { data: members = [] } = useMembers();
  // mark as read on press
  const { mutate: readNotification } = useReadNotificationMutation();
  const actor = members.find((member) => member.id === actorId);
  const isUnread = !readAt;
  const storyTerm = getTermDisplay("storyTerm");

  const handlePress = () => {
    readNotification(id);
    router.push(`/story/${entityId}`);
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
      className="py-3 px-4 active:bg-gray-50 dark:active:bg-dark-200"
      onPress={handlePress}
    >
      <Row align="center" gap={2}>
        <Avatar
          size="lg"
          name={actor?.fullName || actor?.username || "Unknown"}
          src={actor?.avatarUrl}
          className="shrink-0"
        />
        <Col flex={1} className=" gap-0.5">
          <Row justify="between" align="center" gap={2}>
            <Text
              color={isUnread ? "black" : "muted"}
              className="flex-1 mr-2"
              fontWeight={isUnread ? "medium" : undefined}
              numberOfLines={1}
            >
              {title}
            </Text>
            <Text
              fontSize="sm"
              color={isUnread ? "black" : "muted"}
              fontWeight={isUnread ? "medium" : undefined}
              className="shrink-0"
            >
              {formatTimeAgo(createdAt)}
            </Text>
          </Row>
          <Row align="center" justify="between" gap={2}>
            <Text numberOfLines={1} align="center" className="flex-1">
              {renderTemplateJSX(message, storyTerm)}
            </Text>
            {isUnread && <Dot color={colors.primary} size={10} />}
          </Row>
        </Col>
      </Row>
    </Pressable>
  );
};
