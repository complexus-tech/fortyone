"use client";
import { Avatar, Box, ContextMenu, Flex, Text, TimeAgo, Tooltip } from "ui";
import {
  DeleteIcon,
  NotificationsCheckIcon,
  NotificationsUnreadIcon,
} from "icons";
import Link from "next/link";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { ListItemAttentionDot } from "@/components/ui/list-item-attention-dot";
import { MayaAvatar } from "@/components/ui/maya-avatar";
import { useTerminology, useWorkspacePath } from "@/hooks";
import type { AppNotification } from "./types";
import { useReadNotificationMutation } from "./hooks/read-mutation";
import { useMarkUnreadMutation } from "./hooks/mark-unread-mutation";
import { useDeleteMutation } from "./hooks/delete-mutation";
import { NotificationIndicator } from "./notification-indicator";
import { NotificationMessageContent } from "./notification-message-content";
import { renderTemplate } from "./utils/render-template";
import { getNotificationDetailsPath } from "./utils/notification-destination";
import { prepareNotificationMessage } from "./utils/prepare-notification-message";

export const NotificationCard = ({
  id,
  title,
  message,
  type,
  entityId,
  entityType,
  readAt,
  createdAt,
  actor,
}: AppNotification) => {
  const pathname = usePathname();
  const { withWorkspace } = useWorkspacePath();
  const mayaActor = actor?.isSystem ? actor : null;
  const isActive = pathname.includes(id);
  const isUnread = !readAt;
  const { mutate: readNotification } = useReadNotificationMutation();
  const { mutate: unreadNotification } = useMarkUnreadMutation();
  const { mutate: deleteNotification } = useDeleteMutation();
  const { getTermDisplay } = useTerminology();

  const handleReadNotification = () => {
    readNotification({ id, observedCreatedAt: createdAt });
  };

  const handleDelete = () => {
    deleteNotification(id);
  };

  const handleMarkUnread = () => {
    unreadNotification(id);
  };
  const storyTerm = getTermDisplay("storyTerm");
  const renderedMessage = renderTemplate(
    prepareNotificationMessage(message, actor, type),
  );
  const messageContent = (
    <NotificationMessageContent
      segments={renderedMessage.segments}
      storyTerm={storyTerm}
    />
  );

  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Box className="min-w-0">
          <Link
            className="block min-w-0"
            href={withWorkspace(
              getNotificationDetailsPath({
                entityId,
                entityType,
                notificationId: id,
                observedCreatedAt: createdAt,
                message,
              }),
            )}
            prefetch={false}
          >
            <Box
              className={cn(
                "border-border block min-w-0 cursor-pointer overflow-hidden border-b-[0.5px] px-5 py-[0.655rem] transition md:px-4",
                {
                  "bg-primary/5 hover:bg-primary/5": isActive,
                  "hover:bg-surface-muted": !isActive,
                },
              )}
            >
              <Flex align="center" className="mb-2" gap={2} justify="between">
                <Flex align="center" className="min-w-0 flex-1" gap={2}>
                  {isUnread ? <ListItemAttentionDot /> : null}
                  <Text
                    className="line-clamp-1 min-w-0 flex-1 truncate font-medium"
                    color={isUnread ? undefined : "muted"}
                  >
                    {title}
                  </Text>
                </Flex>
                <Text className="shrink-0 text-[0.95rem]" color="muted">
                  <TimeAgo timestamp={createdAt} />
                </Text>
              </Flex>
              <Flex align="center" gap={3} justify="between">
                <Flex align="center" className="min-w-0 flex-1" gap={2}>
                  {mayaActor ? (
                    <MayaAvatar
                      className="shrink-0"
                      name={mayaActor.fullName}
                      size="xs"
                      src={mayaActor.avatarUrl}
                    />
                  ) : (
                    <Avatar
                      className="shrink-0"
                      name={actor?.fullName || actor?.username}
                      size="xs"
                      src={actor?.avatarUrl}
                    />
                  )}

                  <Tooltip
                    className="max-h-[min(24rem,80vh)] max-w-[min(24rem,calc(100vw-2rem))] overflow-y-auto [overflow-wrap:anywhere]"
                    title={
                      <span className="whitespace-pre-wrap">
                        {messageContent}
                      </span>
                    }
                  >
                    <Text
                      className="line-clamp-1 min-w-0 flex-1 truncate"
                      color="muted"
                    >
                      <span>{messageContent}</span>
                    </Text>
                  </Tooltip>
                </Flex>
                <NotificationIndicator message={message} type={type} />
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
