import { Box } from "ui";
import { NotificationCard } from "@/components/notifications/card";
import { NotificationMessage } from "@/components/notifications/message";
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
        {notifications.map(({ id, read }) => (
          <NotificationCard key={id} read={read} />
        ))}
      </Box>
      <NotificationMessage />
    </Box>
  );
};
