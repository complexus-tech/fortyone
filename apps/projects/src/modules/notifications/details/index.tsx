"use client";
import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { StoryPage } from "@/modules/story";
import { notificationKeys } from "@/constants/keys";

export const NotificationDetails = ({
  entityId,
}: {
  notificationId: string;
  entityId: string;
  entityType: "story" | "objective";
}) => {
  const queryClient = useQueryClient();
  useEffect(() => {
    queryClient.invalidateQueries({
      queryKey: notificationKeys.unread(),
    });
    queryClient.invalidateQueries({
      queryKey: notificationKeys.all,
    });
  }, [queryClient]);

  return <StoryPage isNotifications storyId={entityId} />;
};
