"use client";
import { NotificationsIcon } from "icons";
import { Flex, Text } from "ui";
import { useNotifications } from "./hooks/notifications";

export const SelectNotificationMessage = () => {
  const { data: notifications = [] } = useNotifications();
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
