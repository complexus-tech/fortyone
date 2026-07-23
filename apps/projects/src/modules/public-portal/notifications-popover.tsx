"use client";

import type { InfiniteData } from "@tanstack/react-query";
import { useState } from "react";
import Link from "next/link";
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  BellIcon,
  CheckIcon,
  MoreVerticalIcon,
  NotificationsCheckIcon,
  NotificationsUnreadIcon,
} from "icons";
import {
  Avatar,
  Badge,
  Box,
  Button,
  Flex,
  Menu,
  Popover,
  Text,
  TimeAgo,
} from "ui";
import { cn } from "lib";
import { toast } from "sonner";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import {
  getPublicPortalNotificationsAction,
  getPublicPortalUnreadCountAction,
  markAllPublicPortalNotificationsReadAction,
  markPublicPortalNotificationReadAction,
} from "./notification-actions";
import type {
  PublicPortal,
  PublicPortalNotification,
  PublicPortalNotificationsPage,
} from "./types";
import { getPublicAvatarColor } from "./avatar-color";
import { publicPortalKeys } from "./query-keys";
import { getRequestPathBySlug } from "./utils";

const PAGE_SIZE = 10;

const getNotificationPage = async (
  portalSlug: string,
  page: number,
  unreadOnly: boolean,
) => {
  const response = await getPublicPortalNotificationsAction({
    page,
    pageSize: PAGE_SIZE,
    portalSlug,
    unreadOnly,
  });

  if (response.error?.message || !response.data) {
    throw new Error(response.error?.message || "Unable to load notifications");
  }

  return response.data;
};

const getUnreadCount = async (portalSlug: string) => {
  const response = await getPublicPortalUnreadCountAction(portalSlug);

  if (response.error?.message || !response.data) {
    throw new Error(
      response.error?.message || "Unable to load unread notifications",
    );
  }

  return response.data.count;
};

const renderNotificationMessage = (notification: PublicPortalNotification) =>
  notification.message.template.replace(/\{\w+\}/g, (placeholder) => {
    const key = placeholder.slice(1, -1);
    return notification.message.variables[key]?.value ?? placeholder;
  });

const NotificationsSkeleton = () => (
  <Box className="space-y-1 px-2 py-1">
    {Array.from({ length: 4 }).map((_, index) => (
      <Flex className="animate-pulse px-2 py-3" gap={3} key={index}>
        <Box className="bg-surface-muted size-9 shrink-0 rounded-full" />
        <Box className="min-w-0 flex-1">
          <Box className="bg-surface-muted h-4 w-4/5 rounded-md" />
          <Box className="bg-surface-muted mt-2 h-3.5 w-2/5 rounded-md" />
        </Box>
      </Flex>
    ))}
  </Box>
);

const EmptyNotifications = () => (
  <Flex
    align="center"
    className="px-8 py-12 text-center"
    direction="column"
    justify="center"
  >
    <BellIcon className="text-text-muted h-8 w-auto" strokeWidth={1.5} />
    <Text className="mt-4" fontWeight="semibold">
      You&apos;re all caught up
    </Text>
    <Text className="mt-1 max-w-64 text-[0.95rem]" color="muted">
      Updates about feedback you submit will appear here.
    </Text>
  </Flex>
);

const NotificationItem = ({
  isLast,
  notification,
  onRead,
  portal,
}: {
  isLast: boolean;
  notification: PublicPortalNotification;
  onRead: (notification: PublicPortalNotification) => void;
  portal: PublicPortal;
}) => {
  const isUnread = !notification.readAt;

  return (
    <Link
      className={cn(
        "border-border/60 dark:border-border-strong/80 hover:bg-state-hover/40 relative block border-b-[0.5px] px-4 py-3 transition-colors",
        { "border-b-0": isLast },
      )}
      href={getRequestPathBySlug(portal, notification.feedback.slug)}
      onClick={() => {
        onRead(notification);
      }}
    >
      <Flex align="center" gap={3}>
        <Avatar
          className="shrink-0"
          name={notification.actor.name}
          rounded="full"
          size="sm"
          src={notification.actor.avatarUrl}
          style={{
            backgroundColor: getPublicAvatarColor(notification.actor.name),
          }}
        />
        <Box className="min-w-0 flex-1">
          <Flex align="center" gap={2} justify="between">
            <Text
              className="line-clamp-1 min-w-0 flex-1 text-base leading-6"
              fontWeight={isUnread ? "semibold" : "medium"}
            >
              {renderNotificationMessage(notification)}
            </Text>
            {isUnread ? (
              <span className="bg-primary size-2 shrink-0 rounded-full" />
            ) : null}
          </Flex>
          <Flex align="center" className="mt-1" gap={3} justify="between">
            <Text
              className="line-clamp-1 min-w-0 flex-1 text-base leading-6"
              color="muted"
            >
              {notification.feedback.title}
            </Text>
            <Text
              className="shrink-0 text-sm leading-6 whitespace-nowrap"
              color="muted"
            >
              <TimeAgo timestamp={notification.createdAt} />
            </Text>
          </Flex>
        </Box>
      </Flex>
    </Link>
  );
};

export const PublicPortalNotifications = ({
  portal,
}: {
  portal: PublicPortal;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [showUnreadOnly, setShowUnreadOnly] = useState(false);
  const queryClient = useQueryClient();
  const notificationListKey = publicPortalKeys.notificationList(
    portal.slug,
    showUnreadOnly,
  );
  const notificationsQuery = useInfiniteQuery({
    enabled: isOpen,
    queryKey: notificationListKey,
    queryFn: ({ pageParam }) =>
      getNotificationPage(portal.slug, pageParam, showUnreadOnly),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
  });
  const unreadQuery = useQuery({
    queryKey: publicPortalKeys.notificationUnreadCount(portal.slug),
    queryFn: () => getUnreadCount(portal.slug),
    refetchInterval: DURATION_FROM_MILLISECONDS.MINUTE,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
  });
  const notifications =
    notificationsQuery.data?.pages.flatMap((page) => page.notifications) ?? [];
  const unreadCount = unreadQuery.data ?? 0;
  const { fetchNextPage, hasNextPage, isFetchingNextPage } = notificationsQuery;

  const markReadMutation = useMutation({
    mutationFn: async (notificationId: string) => {
      const response = await markPublicPortalNotificationReadAction({
        notificationId,
        portalSlug: portal.slug,
      });
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
    },
    onMutate: async (notificationId) => {
      const unreadKey = publicPortalKeys.notificationUnreadCount(portal.slug);
      await Promise.all([
        queryClient.cancelQueries({ queryKey: notificationListKey }),
        queryClient.cancelQueries({ queryKey: unreadKey }),
      ]);
      const previousNotifications =
        queryClient.getQueryData<
          InfiniteData<PublicPortalNotificationsPage, number>
        >(notificationListKey);
      const previousUnreadCount = queryClient.getQueryData<number>(unreadKey);
      const wasUnread = previousNotifications?.pages.some((page) =>
        page.notifications.some(
          (notification) =>
            notification.id === notificationId && !notification.readAt,
        ),
      );

      queryClient.setQueryData<
        InfiniteData<PublicPortalNotificationsPage, number>
      >(notificationListKey, (current) =>
        current
          ? {
              ...current,
              pages: current.pages.map((page) => ({
                ...page,
                notifications: page.notifications.map((notification) =>
                  notification.id === notificationId
                    ? { ...notification, readAt: new Date().toISOString() }
                    : notification,
                ),
              })),
            }
          : current,
      );
      if (wasUnread && previousUnreadCount !== undefined) {
        queryClient.setQueryData(
          unreadKey,
          Math.max(0, previousUnreadCount - 1),
        );
      }

      return { previousNotifications, previousUnreadCount };
    },
    onError: (error, _notificationId, context) => {
      if (context?.previousNotifications) {
        queryClient.setQueryData(
          notificationListKey,
          context.previousNotifications,
        );
      }
      if (context?.previousUnreadCount !== undefined) {
        queryClient.setQueryData(
          publicPortalKeys.notificationUnreadCount(portal.slug),
          context.previousUnreadCount,
        );
      }
      toast.error("Unable to update notification", {
        description: error.message,
      });
    },
    onSettled: () => {
      void queryClient.invalidateQueries({
        queryKey: publicPortalKeys.notifications(portal.slug),
      });
    },
  });

  const markAllReadMutation = useMutation({
    mutationFn: async () => {
      const response = await markAllPublicPortalNotificationsReadAction(
        portal.slug,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
    },
    onMutate: async () => {
      const listsKey = publicPortalKeys.notificationLists(portal.slug);
      const unreadKey = publicPortalKeys.notificationUnreadCount(portal.slug);
      await Promise.all([
        queryClient.cancelQueries({ queryKey: listsKey }),
        queryClient.cancelQueries({ queryKey: unreadKey }),
      ]);

      const previousLists = queryClient.getQueriesData<
        InfiniteData<PublicPortalNotificationsPage, number>
      >({ queryKey: listsKey });
      const previousUnreadCount = queryClient.getQueryData<number>(unreadKey);
      const readAt = new Date().toISOString();

      previousLists.forEach(([queryKey, current]) => {
        const queryFilter = queryKey.at(-1) as
          | { unreadOnly?: boolean }
          | undefined;
        queryClient.setQueryData(queryKey, (existing: typeof current) => {
          if (!existing) return existing;
          if (queryFilter?.unreadOnly) {
            return {
              ...existing,
              pages: existing.pages.map((page) => ({
                ...page,
                notifications: [],
                pagination: {
                  ...page.pagination,
                  hasMore: false,
                },
              })),
            };
          }
          return {
            ...existing,
            pages: existing.pages.map((page) => ({
              ...page,
              notifications: page.notifications.map((notification) => ({
                ...notification,
                readAt,
              })),
            })),
          };
        });
      });
      queryClient.setQueryData(unreadKey, 0);

      return { previousLists, previousUnreadCount };
    },
    onError: (error, _variables, context) => {
      context?.previousLists.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data);
      });
      if (context?.previousUnreadCount !== undefined) {
        queryClient.setQueryData(
          publicPortalKeys.notificationUnreadCount(portal.slug),
          context.previousUnreadCount,
        );
      }
      toast.error("Unable to mark notifications as read", {
        description: error.message,
      });
    },
    onSuccess: () => {
      toast.success("All notifications marked as read");
    },
    onSettled: () => {
      void queryClient.invalidateQueries({
        queryKey: publicPortalKeys.notifications(portal.slug),
      });
    },
  });

  return (
    <Popover
      onOpenChange={(open) => {
        setIsOpen(open);
      }}
      open={isOpen}
    >
      <Popover.Trigger asChild>
        <Button
          aria-label="Notifications"
          asIcon
          className="relative"
          color="tertiary"
          size="md"
          variant="naked"
        >
          <BellIcon className="h-[1.35rem] w-auto" />
          {unreadCount > 0 ? (
            <Badge
              className="absolute -top-1 -right-1 shrink-0"
              rounded="full"
              size="sm"
            >
              {unreadCount > 9 ? "9+" : unreadCount}
            </Badge>
          ) : null}
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        aria-label="Notifications"
        className="w-[min(28rem,calc(100vw-2rem))] overflow-hidden p-0"
        sideOffset={8}
      >
        <Flex
          align="center"
          className="border-border/60 dark:border-border-strong/80 h-14 border-b-[0.5px] px-4"
          justify="between"
        >
          <Text fontWeight="semibold">Notifications</Text>
          <Flex align="center">
            <Menu>
              <Menu.Button>
                <Button
                  aria-label="Notification actions"
                  asIcon
                  className="text-text-muted"
                  color="tertiary"
                  size="sm"
                  variant="naked"
                >
                  <MoreVerticalIcon />
                </Button>
              </Menu.Button>
              <Menu.Items align="end" className="w-56" sideOffset={8}>
                <Menu.Group>
                  <Menu.Item
                    active={!showUnreadOnly}
                    className="justify-between"
                    onSelect={() => {
                      setShowUnreadOnly(false);
                    }}
                  >
                    <span className="flex items-center gap-2">
                      <BellIcon className="h-5 w-auto" />
                      All notifications
                    </span>
                    {!showUnreadOnly ? (
                      <CheckIcon className="h-4 w-auto" />
                    ) : null}
                  </Menu.Item>
                  <Menu.Item
                    active={showUnreadOnly}
                    className="justify-between"
                    onSelect={() => {
                      setShowUnreadOnly(true);
                    }}
                  >
                    <span className="flex items-center gap-2">
                      <NotificationsUnreadIcon className="h-5 w-auto" />
                      Unread only
                    </span>
                    {showUnreadOnly ? (
                      <CheckIcon className="h-4 w-auto" />
                    ) : null}
                  </Menu.Item>
                </Menu.Group>
                <Menu.Separator />
                <Menu.Group>
                  <Menu.Item
                    disabled={
                      unreadCount === 0 || markAllReadMutation.isPending
                    }
                    onSelect={() => {
                      markAllReadMutation.mutate();
                    }}
                  >
                    <NotificationsCheckIcon className="h-5 w-auto" />
                    Mark all as read
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
        </Flex>
        <Box className="hide-scrollbar max-h-[min(30rem,calc(100dvh-7rem))] overflow-y-auto py-1">
          {notificationsQuery.isPending ? <NotificationsSkeleton /> : null}
          {notificationsQuery.isError ? (
            <Flex
              align="center"
              className="px-6 py-12 text-center"
              direction="column"
            >
              <Text fontWeight="semibold">Unable to load notifications</Text>
              <Text className="mt-1 text-sm" color="muted">
                Check your connection and try again.
              </Text>
              <Button
                className="mt-4"
                color="tertiary"
                onClick={() => {
                  void notificationsQuery.refetch();
                }}
                size="sm"
              >
                Try again
              </Button>
            </Flex>
          ) : null}
          {!notificationsQuery.isPending &&
          !notificationsQuery.isError &&
          notifications.length === 0 ? (
            <EmptyNotifications />
          ) : null}
          {notifications.map((notification, index) => (
            <NotificationItem
              isLast={index === notifications.length - 1}
              key={notification.id}
              notification={notification}
              onRead={(selectedNotification) => {
                setIsOpen(false);
                if (!selectedNotification.readAt) {
                  markReadMutation.mutate(selectedNotification.id);
                }
              }}
              portal={portal}
            />
          ))}
          {hasNextPage ? (
            <Flex
              align="center"
              className="border-border/60 dark:border-border-strong/80 border-t-[0.5px] px-3 py-2"
              justify="center"
            >
              <Button
                color="tertiary"
                disabled={isFetchingNextPage}
                onClick={() => {
                  void fetchNextPage();
                }}
                size="sm"
                variant="naked"
              >
                {isFetchingNextPage ? "Loading..." : "Load more notifications"}
              </Button>
            </Flex>
          ) : null}
        </Box>
      </Popover.Content>
    </Popover>
  );
};
