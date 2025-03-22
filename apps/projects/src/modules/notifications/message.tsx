"use client";
import { NotificationsIcon } from "icons";
import { Flex, Text } from "ui";
import { useNotifications } from "./hooks/notifications";

const SelectNotificationSkeleton = () => {
  return (
    <Flex align="center" className="h-full" justify="center">
      <Flex align="center" direction="column">
        <div className="mb-3 h-16 w-16 animate-pulse rounded-full bg-gray-100 dark:bg-dark-200" />
        <div className="h-5 w-40 animate-pulse rounded bg-gray-100 dark:bg-dark-200" />
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
        <NotificationsIcon className="mb-3 h-16 w-auto" />
        {notifications.length === 0 ? (
          <Text color="muted">You have no notifications</Text>
        ) : (
          <Text color="muted">Select a notification to view</Text>
        )}
      </Flex>
    </Flex>
  );
};
