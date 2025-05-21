"use client";
import { useEffect, useRef } from "react";
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
  const hasMounted = useRef(false);
  const { mutate: readNotification } = useReadNotificationMutation(false);

  useEffect(() => {
    if (!hasMounted.current) {
      readNotification(notificationId);
      hasMounted.current = true;
    }
  }, [notificationId, readNotification]);

  return <StoryPage isNotifications storyId={entityId} />;
};
