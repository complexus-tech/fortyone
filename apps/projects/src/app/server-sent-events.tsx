"use client";

import { useQueryClient } from "@tanstack/react-query";
import { usePostHog } from "posthog-js/react";
import { useCallback, useEffect } from "react";
import { getApiUrl } from "@/lib/api-url";
import type { AppNotification } from "@/modules/notifications/types";
import { storyKeys } from "@/modules/stories/constants";
import { notificationKeys } from "@/constants/keys";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useWorkspacePath } from "@/hooks";
import type { DetailedStory } from "@/modules/story/types";
import type { Story } from "@/modules/stories/types";

const apiURL = getApiUrl();

// NEW: Type for workspace updates
type WorkspaceUpdate = {
  type: "story.workspace_update";
  storyId: string;
  workspaceId: string;
  changes: {
    statusId?: string;
    assigneeId?: string;
    priority?: string;
    title?: string;
  };
  actorId: string;
  actorName: string;
  timestamp: number;
};

export const ServerSentEvents = () => {
  const posthog = usePostHog();
  const { workspace } = useCurrentWorkspace();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();

  const handleNotification = useCallback(
    (notification: AppNotification) => {
      queryClient.invalidateQueries({
        queryKey: notificationKeys.all(workspaceSlug),
      });
      if (notification.entityType === "story") {
        queryClient.invalidateQueries({
          queryKey: storyKeys.detail(workspaceSlug, notification.entityId),
        });
      }
    },
    [queryClient, workspaceSlug],
  );

  const handleWorkspaceUpdate = useCallback(
    (workspaceUpdate: WorkspaceUpdate) => {
      queryClient.setQueriesData(
        {
          predicate: (query) => {
            const queryKey = query.queryKey;
            return (
              Array.isArray(queryKey) &&
              queryKey[0] === "stories" &&
              !queryKey.includes("detail") &&
              query.isActive()
            );
          },
        },
        (oldData: Story[] | undefined) => {
          if (!oldData) return oldData;
          // Find and update the specific story
          return oldData.map((story) =>
            story.id === workspaceUpdate.storyId
              ? {
                  ...story,
                  ...workspaceUpdate.changes,
                }
              : story,
          );
        },
      );

      queryClient.setQueryData(
        storyKeys.detail(workspaceSlug, workspaceUpdate.storyId),
        (oldData: DetailedStory | undefined) => {
          if (!oldData) return oldData;
          return {
            ...oldData,
            ...workspaceUpdate.changes,
          };
        },
      );
    },
    [queryClient, workspaceSlug],
  );

  useEffect(() => {
    if (!workspace?.slug) return;

    const SSE_ENDPOINT = `${apiURL}/workspaces/${workspace.slug}/notifications/subscribe`;
    const eventSource = new EventSource(SSE_ENDPOINT, {
      withCredentials: true,
    });

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(`${event.data}`);

        if (data.type === "story.workspace_update") {
          const workspaceUpdate = data as WorkspaceUpdate;
          handleWorkspaceUpdate(workspaceUpdate);
        } else {
          const notification = data as AppNotification;
          handleNotification(notification);
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
  }, [posthog, workspace?.slug, handleNotification, handleWorkspaceUpdate]);

  return null;
};
