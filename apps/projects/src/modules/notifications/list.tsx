"use client";
import { useEffect } from "react";
import { parseAsString, useQueryState } from "nuqs";
import { useIntersectionObserver } from "react-intersection-observer-hook";
import { Box, Flex, Skeleton, Text } from "ui";
import { NotificationCard } from "@/modules/notifications/card";
import { useTerminology } from "@/hooks";
import { NotificationsHeader } from "./header";
import { useNotificationsInfinite } from "./hooks/notifications";
import { NotificationsSkeleton } from "./notifications-skeleton";

export const ListNotifications = () => {
  const { getTermDisplay } = useTerminology();
  const [search, setSearch] = useQueryState(
    "search",
    parseAsString.withDefault(""),
  );
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending } =
    useNotificationsInfinite(search);
  const notifications = data?.pages.flatMap((page) => page.notifications) ?? [];
  const [triggerRef, { entry }] = useIntersectionObserver({
    threshold: 0,
    rootMargin: "240px",
  });

  useEffect(() => {
    if (entry?.isIntersecting && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [entry?.isIntersecting, fetchNextPage, hasNextPage, isFetchingNextPage]);

  if (isPending) return <NotificationsSkeleton />;
  return (
    <Box className="border-border/60 d h-dvh border-r-[0.5px] pb-6">
      <NotificationsHeader
        onSearchChange={(nextSearch) => {
          void setSearch(nextSearch || null);
        }}
        search={search}
      />
      <Box className="h-[calc(100dvh-4rem)] overflow-y-auto">
        {notifications.map((notification, idx) => (
          <NotificationCard
            key={notification.id}
            {...notification}
            index={idx}
          />
        ))}
        {hasNextPage ? <div className="h-4 w-full" ref={triggerRef} /> : null}
        {isFetchingNextPage ? (
          <Box>
            {Array.from({ length: 3 }).map((_, index) => (
              <Flex
                className="border-border/70 d/60 border-b px-4 py-3"
                direction="column"
                gap={2}
                key={index}
              >
                <Flex align="center" justify="between">
                  <Skeleton className="h-4 w-3/5" />
                  <Skeleton className="h-4 w-10" />
                </Flex>
                <Flex align="center" justify="between">
                  <Flex align="center" gap={2}>
                    <Skeleton className="size-8 rounded-full" />
                    <Skeleton className="h-4 w-40" />
                  </Flex>
                  <Skeleton className="size-4" />
                </Flex>
              </Flex>
            ))}
          </Box>
        ) : null}
        {notifications.length === 0 && (
          <Flex align="center" className="h-full px-6" justify="center">
            <Box>
              <Text align="center" className="mb-3" fontSize="xl">
                {search ? "No matching notifications" : "No notifications"}
              </Text>
              <Text align="center" color="muted">
                {search
                  ? `No notifications match “${search}”.`
                  : `You will receive notifications when you are assigned or mentioned in a ${getTermDisplay("storyTerm")}.`}
              </Text>
            </Box>
          </Flex>
        )}
      </Box>
    </Box>
  );
};
