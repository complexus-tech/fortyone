"use client";
import { Avatar, Box, ContextMenu, Flex, Text, TimeAgo, Tooltip } from "ui";
import {
  AtIcon,
  CalendarIcon,
  CommentIcon,
  DeleteIcon,
  NotificationsCheckIcon,
  NotificationsUnreadIcon,
} from "icons";
import Link from "next/link";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { PriorityIcon, StoryStatusIcon } from "@/components/ui";
import { ListItemAttentionDot } from "@/components/ui/list-item-attention-dot";
import { MayaAvatar } from "@/components/ui/maya-avatar";
import { useTerminology, useWorkspacePath } from "@/hooks";
import type { AppNotification } from "./types";
import { useReadNotificationMutation } from "./hooks/read-mutation";
import { useMarkUnreadMutation } from "./hooks/mark-unread-mutation";
import { useDeleteMutation } from "./hooks/delete-mutation";
import { NotificationMessageContent } from "./notification-message-content";
import { renderTemplate } from "./utils/render-template";
import { getNotificationDetailsPath } from "./utils/notification-destination";

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
  const hasActorVariable = Object.hasOwn(message.variables, "actor");
  const messageWithActor =
    actor && !hasActorVariable
      ? {
          ...message,
          template: `{actor} ${message.template}`,
          variables: {
            ...message.variables,
            actor: {
              value: actor.fullName || actor.username || "Someone",
              type: "actor",
            },
          },
        }
      : message;
  const renderedMessage = renderTemplate(messageWithActor);
  const text = renderedMessage.text.replace("story", storyTerm);
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
                "border-border block min-w-0 cursor-pointer overflow-hidden border-b-[0.5px] px-5 py-4 text-base leading-normal transition md:px-4",
                {
                  "bg-primary/5 hover:bg-primary/5": isActive,
                  "hover:bg-surface-muted": !isActive,
                },
              )}
            >
              <Flex align="start" gap={3}>
                {mayaActor ? (
                  <MayaAvatar
                    className="size-8 shrink-0 text-base leading-none"
                    name={mayaActor.fullName}
                    size="md"
                    src={mayaActor.avatarUrl}
                  />
                ) : (
                  <Avatar
                    className="size-8 shrink-0 text-base leading-none"
                    name={actor?.fullName || actor?.username}
                    size="md"
                    src={actor?.avatarUrl}
                  />
                )}
                <Box className="min-w-0 flex-1">
                  <Flex
                    align="center"
                    className="mb-1 flex-wrap gap-x-3 gap-y-1"
                    justify="between"
                  >
                    <Flex align="center" className="min-w-0" gap={2}>
                      {isUnread ? <ListItemAttentionDot /> : null}
                      <Text
                        className="min-w-0 [overflow-wrap:anywhere]"
                        color={isUnread ? undefined : "muted"}
                        fontWeight="semibold"
                      >
                        {title}
                      </Text>
                    </Flex>
                    <Text className="shrink-0" color="muted">
                      <TimeAgo timestamp={createdAt} />
                    </Text>
                  </Flex>
                  <Flex align="start" gap={3} justify="between">
                    <Tooltip
                      className="max-h-[min(24rem,80vh)] max-w-[min(24rem,calc(100vw-2rem))] overflow-y-auto text-base leading-normal [overflow-wrap:anywhere]"
                      title={
                        <span className="whitespace-pre-wrap">
                          {messageContent}
                        </span>
                      }
                    >
                      <Text
                        className="line-clamp-2 min-w-0 flex-1 [overflow-wrap:anywhere]"
                        color="muted"
                      >
                        <span>{messageContent}</span>
                      </Text>
                    </Tooltip>
                    {type === "story_update" && (
                      <>
                        {text.toLowerCase().includes("deadline") && (
                          <CalendarIcon className="shrink-0" />
                        )}
                        {text.toLowerCase().includes("status") && (
                          <StoryStatusIcon className="shrink-0" />
                        )}
                        {text.toLowerCase().includes("priority") && (
                          <PriorityIcon
                            className="text-foreground shrink-0"
                            priority="High"
                          />
                        )}
                      </>
                    )}
                    {type === "story_comment" && (
                      <CommentIcon className="shrink-0" />
                    )}
                    {type === "mention" && <AtIcon className="shrink-0" />}
                  </Flex>
                </Box>
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
