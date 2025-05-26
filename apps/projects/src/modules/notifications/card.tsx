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
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { useMembers } from "@/lib/hooks/members";
import type { AppNotification } from "./types";
import { useReadNotificationMutation } from "./hooks/read-mutation";
import { useMarkUnreadMutation } from "./hooks/mark-unread-mutation";
import { useDeleteMutation } from "./hooks/delete-mutation";
import { renderTemplate } from "./utils/render-template";

export const NotificationCard = ({
  id,
  title,
  message,
  entityType,
  entityId,
  readAt,
  createdAt,
  actorId,
  index,
}: AppNotification & { index: number }) => {
  const pathname = usePathname();
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
            href={`/notifications/${id}?entityId=${entityId}&entityType=${entityType}`}
            prefetch={index <= 10 ? true : null}
          >
            <Box
              className={cn(
                "block cursor-pointer border-b-[0.5px] border-gray-100 px-5 py-[0.655rem] transition hover:bg-gray-50/60 dark:border-dark-100 dark:hover:bg-dark-300/90 md:px-4",
                {
                  "bg-gray-100/70 dark:bg-dark-200/80": pathname.includes(id),
                  "relative left-px border-l-[1.5px] border-l-primary dark:border-l-primary":
                    isUnread,
                },
              )}
            >
              <Flex align="center" className="mb-2" gap={2} justify="between">
                <Text
                  className="line-clamp-1 flex-1 font-medium"
                  color={isUnread ? undefined : "muted"}
                >
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

                  <Tooltip
                    className="max-w-[200px]"
                    title={renderTemplate(message).text}
                  >
                    <Text className="line-clamp-1" color="muted">
                      <span
                        dangerouslySetInnerHTML={{
                          __html: renderTemplate(message).html,
                        }}
                      />
                    </Text>
                  </Tooltip>
                </Flex>
                {entityType === "story" && <StoryIcon className="shrink-0" />}
                {entityType === "objective" && (
                  <ObjectiveIcon className="shrink-0" />
                )}
              </Flex>
            </Box>
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
