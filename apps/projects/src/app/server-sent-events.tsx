"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { usePostHog } from "posthog-js/react";
import { useEffect } from "react";
import type { AppNotification } from "@/modules/notifications/types";
import { storyKeys } from "@/modules/stories/constants";
import { notificationKeys } from "@/constants/keys";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";

const apiURL = process.env.NEXT_PUBLIC_API_URL!;

export const ServerSentEvents = () => {
  const posthog = usePostHog();
  const { data: session } = useSession();
  const { workspace } = useCurrentWorkspace();
  const queryClient = useQueryClient();

  useEffect(() => {
    const SSE_ENDPOINT = `${apiURL}/workspaces/${workspace?.id}/notifications/subscribe?token=${session?.token}`;
    const eventSource = new EventSource(SSE_ENDPOINT);

    eventSource.onopen = () => {
      queryClient.invalidateQueries({
        queryKey: notificationKeys.all,
      });
    };

    eventSource.onmessage = (event) => {
      try {
        const notification = JSON.parse(`${event.data}`) as AppNotification;
        queryClient.invalidateQueries({
          queryKey: notificationKeys.all,
        });
        if (notification.entityType === "story") {
          queryClient.invalidateQueries({
            queryKey: storyKeys.detail(notification.entityId),
          });
          queryClient.invalidateQueries({
            queryKey: storyKeys.mine(),
          });
        }
      } catch (error) {
        posthog.captureException(error);
      }
    };

    eventSource.onerror = (error) => {
      posthog.captureException(error);
    };

    return () => {
      eventSource.close();
    };
  }, [posthog, session?.token, queryClient]);

  return null;
};
