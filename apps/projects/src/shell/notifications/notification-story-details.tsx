"use client";

import { useEffect, useRef } from "react";
import { useReadNotificationMutation } from "@/modules/notifications/public/client";
import { StoryPage } from "@/modules/story/public/client";

type NotificationStoryDetailsProps = {
  entityId: string;
  notificationId: string;
};

/**
 * Route-level composition for a notification that opens a story.
 *
 * The shell coordinates the two feature capabilities so neither feature needs
 * to import the other just to render this route.
 */
export const NotificationStoryDetails = ({
  entityId,
  notificationId,
}: NotificationStoryDetailsProps) => {
  const hasMounted = useRef(false);
  const { mutate: readNotification } = useReadNotificationMutation(false);

  useEffect(() => {
    if (hasMounted.current) return;

    readNotification(notificationId);
    hasMounted.current = true;
  }, [notificationId, readNotification]);

  return <StoryPage isNotifications storyId={entityId} />;
};
