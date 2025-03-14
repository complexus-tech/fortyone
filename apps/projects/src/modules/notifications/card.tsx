"use client";
import { Avatar, Box, ContextMenu, Flex, Text, TimeAgo, Tooltip } from "ui";
import {
  DeleteIcon,
  NotificationsCheckIcon,
  NotificationsUnreadIcon,
  ObjectiveIcon,
  StoryIcon,
} from "icons";
import Link from "next/link";
import { Dot, RowWrapper } from "@/components/ui";
import { useMembers } from "@/lib/hooks/members";
import type { AppNotification } from "./types";
import { useReadNotificationMutation } from "./hooks/read-mutation";
import { useMarkUnreadMutation } from "./hooks/mark-unread-mutation";
import { useDeleteMutation } from "./hooks/delete-mutation";

export const NotificationCard = ({
  id,
  title,
  description,
  entityType,
  entityId,
  readAt,
  createdAt,
  actorId,
}: AppNotification) => {
  const { data: members = [] } = useMembers();
  const actor = members.find((member) => member.id === actorId);
  const isUnread = !readAt;
  const { mutate: readNotification } = useReadNotificationMutation();
  const { mutate: unreadNotification } = useMarkUnreadMutation();
  const { mutate: deleteNotification } = useDeleteMutation();

  const handleReadNotification = () => {
    readNotification(id);
  };

  const handleDelete = () => {
    deleteNotification(id);
  };

  const handleMarkUnread = () => {
    unreadNotification(id);
  };

  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Box>
          <Link
            className="block"
            href={`?entityId=${entityId}&entityType=${entityType}`}
          >
            <RowWrapper className="block cursor-pointer px-4">
              <Flex align="center" className="mb-2" gap={2} justify="between">
                <Text className="line-clamp-1 flex-1 font-medium">
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
                  <Tooltip className="max-w-[300px]" title={description}>
                    <Text className="line-clamp-1" color="muted">
                      {description}
                    </Text>
                  </Tooltip>
                </Flex>
                {entityType === "story" && <StoryIcon className="shrink-0" />}
                {entityType === "objective" && (
                  <ObjectiveIcon className="shrink-0" />
                )}
              </Flex>
            </RowWrapper>
          </Link>
        </Box>
      </ContextMenu.Trigger>
      <ContextMenu.Items>
        <ContextMenu.Group>
          {isUnread ? (
            <ContextMenu.Item onSelect={handleReadNotification}>
              <NotificationsCheckIcon />
              Mark as read
            </ContextMenu.Item>
          ) : (
            <ContextMenu.Item onSelect={handleMarkUnread}>
              <NotificationsUnreadIcon />
              Mark as unread
            </ContextMenu.Item>
          )}
          <ContextMenu.Item onSelect={handleDelete}>
            <DeleteIcon />
            Delete...
          </ContextMenu.Item>
        </ContextMenu.Group>
      </ContextMenu.Items>
    </ContextMenu>
  );
};
