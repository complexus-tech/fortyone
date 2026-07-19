"use client";

import type { InfiniteData } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { BellIcon } from "icons";
import { Avatar, Box, Button, Flex, Popover, Text, TimeAgo } from "ui";
import { cn } from "lib";
import { toast } from "sonner";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import {
  getPublicPortalNotificationsAction,
  getPublicPortalUnreadCountAction,
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

const PAGE_SIZE = 20;

const getNotificationPage = async (portalSlug: string, page: number) => {
  const response = await getPublicPortalNotificationsAction({
    page,
    pageSize: PAGE_SIZE,
    portalSlug,
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
  <Flex align="center" className="px-8 py-14 text-center" direction="column">
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
  notification,
  onRead,
  portal,
}: {
  notification: PublicPortalNotification;
  onRead: (notification: PublicPortalNotification) => void;
  portal: PublicPortal;
}) => {
  const isUnread = !notification.readAt;

  return (
    <Link
      className={cn(
        "hover:bg-surface-muted/60 relative block px-4 py-3 transition-colors",
        { "bg-surface-muted/30": isUnread },
      )}
      href={getRequestPathBySlug(portal, notification.feedback.slug)}
      onClick={() => {
        onRead(notification);
      }}
    >
      <Flex align="start" gap={3}>
        <Avatar
          className="mt-0.5 shrink-0"
          name={notification.actor.name}
          rounded="full"
          size="sm"
          src={notification.actor.avatarUrl}
          style={{
            backgroundColor: getPublicAvatarColor(notification.actor.name),
          }}
        />
        <Box className="min-w-0 flex-1">
          <Flex align="start" gap={2} justify="between">
            <Text
              className="line-clamp-2 text-[0.95rem] leading-5"
              fontWeight={isUnread ? "semibold" : "medium"}
            >
              {renderNotificationMessage(notification)}
            </Text>
            {isUnread ? (
              <span className="bg-primary mt-1.5 size-2 shrink-0 rounded-full" />
            ) : null}
          </Flex>
          <Text className="mt-1 line-clamp-1 text-sm" color="muted">
            {notification.feedback.title}
          </Text>
          <Text className="mt-1 text-xs" color="muted">
            <TimeAgo timestamp={notification.createdAt} />
          </Text>
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
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const queryClient = useQueryClient();
  const notificationsQuery = useInfiniteQuery({
    enabled: isOpen,
    queryKey: publicPortalKeys.notificationList(portal.slug),
    queryFn: ({ pageParam }) => getNotificationPage(portal.slug, pageParam),
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
      const listKey = publicPortalKeys.notificationList(portal.slug);
      const unreadKey = publicPortalKeys.notificationUnreadCount(portal.slug);
      await Promise.all([
        queryClient.cancelQueries({ queryKey: listKey }),
        queryClient.cancelQueries({ queryKey: unreadKey }),
      ]);
      const previousNotifications =
        queryClient.getQueryData<
          InfiniteData<PublicPortalNotificationsPage, number>
        >(listKey);
      const previousUnreadCount = queryClient.getQueryData<number>(unreadKey);
      const wasUnread = previousNotifications?.pages.some((page) =>
        page.notifications.some(
          (notification) =>
            notification.id === notificationId && !notification.readAt,
        ),
      );

      queryClient.setQueryData<
        InfiniteData<PublicPortalNotificationsPage, number>
      >(listKey, (current) =>
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
          publicPortalKeys.notificationList(portal.slug),
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

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!isOpen || !sentinel || !hasNextPage) return;

    const observer = new IntersectionObserver((entries) => {
      if (entries[0]?.isIntersecting && !isFetchingNextPage) {
        void fetchNextPage();
      }
    });
    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [fetchNextPage, hasNextPage, isFetchingNextPage, isOpen]);

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
            <span className="bg-danger border-background absolute top-1.5 right-1.5 flex min-h-2 min-w-2 rounded-full border" />
          ) : null}
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        aria-label="Notifications"
        className="w-[min(24rem,calc(100vw-2rem))] overflow-hidden p-0"
        sideOffset={8}
      >
        <Flex
          align="center"
          className="border-border/60 h-14 border-b-[0.5px] px-4"
          justify="between"
        >
          <Text fontWeight="semibold">Notifications</Text>
          {unreadCount > 0 ? (
            <Text className="text-sm tabular-nums" color="muted">
              {unreadCount} unread
            </Text>
          ) : null}
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
          {notifications.map((notification) => (
            <NotificationItem
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
          <div ref={sentinelRef} />
          {notificationsQuery.isFetchingNextPage ? (
            <NotificationsSkeleton />
          ) : null}
        </Box>
      </Popover.Content>
    </Popover>
  );
};
