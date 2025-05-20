"use client";
import { useEffect } from "react";
import { StoryPage } from "@/modules/story";
import { useReadNotificationMutation } from "@/modules/notifications/hooks/read-mutation";

export const NotificationDetails = ({
  entityId,
  notificationId,
}: {
  notificationId: string;
  entityId: string;
  entityType: "story" | "objective";
}) => {
  const { mutate: readNotification } = useReadNotificationMutation();

  useEffect(() => {
    readNotification(notificationId);
  }, [notificationId, readNotification]);

  return <StoryPage isNotifications storyId={entityId} />;
};
