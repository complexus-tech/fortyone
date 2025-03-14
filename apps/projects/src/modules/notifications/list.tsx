"use client";
import { Box, Flex, Text } from "ui";
import { useSearchParams } from "next/navigation";
import { NotificationCard } from "@/modules/notifications/card";
import { NotificationMessage } from "@/modules/notifications/message";
// import { StoryPage } from "@/modules/story";
import { NotificationsHeader } from "./header";
import { useNotifications } from "./hooks/notifications";

export const ListNotifications = () => {
  const searchParams = useSearchParams();
  const entityId = searchParams.get("entityId");
  const entityType = searchParams.get("entityType");
  const { data: notifications = [] } = useNotifications();

  return (
    <Box className="grid grid-cols-[320px_auto]">
      <Box className="h-screen border-r-[0.5px] border-gray-200/60 pb-6 dark:border-dark-50">
        <NotificationsHeader />
        <Box className="h-[calc(100vh-4rem)] overflow-y-auto">
          {notifications.map((notification) => (
            <NotificationCard key={notification.id} {...notification} />
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
      {entityId && entityType ? (
        <>{entityType === "story" && <Box>{entityId}</Box>}</>
      ) : (
        <NotificationMessage count={notifications.length} />
      )}
    </Box>
  );
};
