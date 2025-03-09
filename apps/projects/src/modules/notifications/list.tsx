import { Box, Flex, Text } from "ui";
import { NotificationCard } from "@/modules/notifications/card";
import { NotificationMessage } from "@/modules/notifications/message";
import { NotificationsHeader } from "./header";

export const ListNotifications = ({
  notifications,
}: {
  notifications: { id: number; read: boolean }[];
}) => {
  return (
    <Box className="grid h-full grid-cols-[350px_auto]">
      <Box className="h-full overflow-y-auto border-r border-gray-100/70 pb-6 dark:border-dark-200">
        <NotificationsHeader />
        <Box className="h-[calc(100vh-4rem)] overflow-y-auto">
          {notifications.map(({ id, read }) => (
            <NotificationCard key={id} read={read} />
          ))}
          {notifications.length === 0 && (
            <Flex align="center" className="h-full px-6" justify="center">
              <Box>
                <Text align="center" className="mb-3" fontSize="xl">
                  No notifications
                </Text>
                <Text align="center" color="muted">
                  You will receive notifications when you are assigned or
                  mentioned in a story.
                </Text>
              </Box>
            </Flex>
          )}
        </Box>
      </Box>
      <NotificationMessage />
    </Box>
  );
};
