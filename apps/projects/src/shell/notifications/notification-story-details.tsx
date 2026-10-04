"use client";

import { useEffect, useRef } from "react";
import { useReadNotificationMutation } from "@/modules/notifications/public/client";
import { StoryPage } from "@/modules/story/public/client";

type NotificationStoryDetailsProps = {
  entityId: string;
  notificationId: string;
  observedCreatedAt?: string;
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
  observedCreatedAt,
}: NotificationStoryDetailsProps) => {
  const readVersion = useRef<string | null>(null);
  const { mutate: readNotification } = useReadNotificationMutation(false);

  useEffect(() => {
    const version = `${notificationId}:${observedCreatedAt ?? ""}`;
    if (readVersion.current === version) return;

    readNotification({ id: notificationId, observedCreatedAt });
    readVersion.current = version;
  }, [notificationId, observedCreatedAt, readNotification]);

  return <StoryPage isNotifications storyId={entityId} />;
};
