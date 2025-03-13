"use client";
import { Avatar, Flex, Text, TimeAgo } from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
import { Dot, RowWrapper } from "@/components/ui";
import { useMembers } from "@/lib/hooks/members";
import type { AppNotification } from "./types";

export const NotificationCard = ({
  title,
  description,
  entityType,
  readAt,
  createdAt,
  actorId,
}: AppNotification) => {
  const { data: members = [] } = useMembers();
  const actor = members.find((member) => member.id === actorId);
  const isUnread = !readAt;

  return (
    <RowWrapper className="block cursor-pointer px-4">
      <Flex align="center" className="mb-2" gap={2} justify="between">
        <Text className="line-clamp-1 flex-1 opacity-90">
          {isUnread ? (
            <Dot className="relative -top-px mr-1 inline size-2.5 shrink-0 items-start text-primary" />
          ) : null}
          {title}
        </Text>
        <Text className="shrink-0" color="muted">
          <TimeAgo timestamp={createdAt} />
        </Text>
      </Flex>

      <Flex align="center" gap={3} justify="between">
        <Flex align="center" className="flex-1" gap={2}>
          <Avatar
            className="shrink-0"
            name={actor?.fullName || actor?.username}
            size="xs"
            src={actor?.avatarUrl}
          />
          <Text className="line-clamp-1" color="muted">
            {description}
          </Text>
        </Flex>
        {entityType === "story" && <StoryIcon className="shrink-0" />}
        {entityType === "objective" && <ObjectiveIcon className="shrink-0" />}
      </Flex>
    </RowWrapper>
  );
};
