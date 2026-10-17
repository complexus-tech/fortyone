"use client";
import { useEffect } from "react";
import { parseAsString, useQueryState } from "nuqs";
import { useIntersectionObserver } from "react-intersection-observer-hook";
import { Box, Button, Flex, Text } from "ui";
import { NotificationCard } from "@/modules/notifications/card";
import { NotificationsEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { useTerminology } from "@/hooks";
import { NotificationsHeader } from "./header";
import { useNotificationsInfinite } from "./hooks/notifications";
import {
  NotificationCardSkeleton,
  NotificationsSkeleton,
} from "./notifications-skeleton";

export const ListNotifications = () => {
  const { getTermDisplay } = useTerminology();
  const [search, setSearch] = useQueryState(
    "search",
    parseAsString.withDefault(""),
  );
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isError,
    isFetchNextPageError,
    isFetching,
    isFetchingNextPage,
    isPending,
    refetch,
  } = useNotificationsInfinite(search);
  const notifications = data?.pages.flatMap((page) => page.notifications) ?? [];
  const [triggerRef, { entry }] = useIntersectionObserver({
    threshold: 0,
    rootMargin: "240px",
  });

  useEffect(() => {
    if (entry?.isIntersecting && hasNextPage && !isFetching && !isError) {
      void fetchNextPage();
    }
  }, [entry?.isIntersecting, fetchNextPage, hasNextPage, isError, isFetching]);

  if (isPending) return <NotificationsSkeleton />;
  return (
    <Box className="border-border flex h-full min-h-0 min-w-0 flex-col overflow-hidden border-r-[0.5px] pb-6">
      <NotificationsHeader
        onSearchChange={(nextSearch) => {
          void setSearch(nextSearch || null);
        }}
        search={search}
      />
      <Box className="min-h-0 flex-1 overflow-y-auto">
        {notifications.map((notification) => (
          <NotificationCard key={notification.id} {...notification} />
        ))}
        {hasNextPage && !isError ? (
          <div className="h-4 w-full" ref={triggerRef} />
        ) : null}
        {isFetchingNextPage ? (
          <Box>
            {Array.from({ length: 3 }).map((_, index) => (
              <NotificationCardSkeleton key={index} />
            ))}
          </Box>
        ) : null}
        {isError ? (
          <Flex
            align="center"
            className={notifications.length ? "px-6 py-8" : "h-full px-6"}
            direction="column"
            gap={3}
            justify="center"
            role="alert"
          >
            <Text align="center" color="muted">
              {isFetchNextPageError
                ? "Couldn't load more notifications."
                : "Couldn't load notifications."}
            </Text>
            <Button
              color="tertiary"
              disabled={isFetching}
              onClick={() => {
                if (isFetchNextPageError) {
                  void fetchNextPage();
                } else {
                  void refetch();
                }
              }}
              size="sm"
              variant="outline"
            >
              {isFetching ? "Retrying..." : "Try again"}
            </Button>
          </Flex>
        ) : null}
        {!isError && notifications.length === 0 && (
          <Flex align="center" className="h-full px-6" justify="center">
            <Box className="flex flex-col items-center">
              <NotificationsEmptyIllustration className="mb-5 w-52" />
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
