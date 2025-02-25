import { Box } from "ui";
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
        <Box className="hidden">
          {notifications.map(({ id, read }) => (
            <NotificationCard key={id} read={read} />
          ))}
        </Box>
      </Box>
      <NotificationMessage />
    </Box>
  );
};
