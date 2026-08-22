"use client";
import { NotificationsIcon } from "icons";
import { Flex, Text } from "ui";
import { NotificationsEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { useNotifications } from "./hooks/notifications";

export const SelectNotificationSkeleton = () => {
  return (
    <Flex align="center" className="h-full" justify="center">
      <Flex align="center" direction="column">
        <div className="bg-skeleton mb-3 h-16 w-16 animate-pulse rounded-full" />
        <div className="bg-skeleton h-5 w-40 animate-pulse rounded" />
      </Flex>
    </Flex>
  );
};

export const SelectNotificationMessage = () => {
  const { data: notifications = [], isPending } = useNotifications();

  if (isPending) return <SelectNotificationSkeleton />;

  return (
    <Flex align="center" className="h-full" justify="center">
      <Flex align="center" direction="column">
        {notifications.length === 0 ? (
          <>
            <NotificationsEmptyIllustration className="mb-4 w-52" />
            <Text color="muted">You have no notifications</Text>
          </>
        ) : (
          <>
            <NotificationsIcon className="mb-3 h-16 w-auto" />
            <Text color="muted">Select a notification to view</Text>
          </>
        )}
      </Flex>
    </Flex>
  );
};
