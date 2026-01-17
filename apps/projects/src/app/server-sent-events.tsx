"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { usePostHog } from "posthog-js/react";
import { useCallback, useEffect } from "react";
import type { AppNotification } from "@/modules/notifications/types";
import { storyKeys } from "@/modules/stories/constants";
import { notificationKeys } from "@/constants/keys";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useWorkspacePath } from "@/hooks";
import type { DetailedStory } from "@/modules/story/types";
import type { Story } from "@/modules/stories/types";

const apiURL = process.env.NEXT_PUBLIC_API_URL!;

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
  const { data: session } = useSession();
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
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        // Only target story list queries that are active
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail") &&
          query.isActive()
        ) {
          queryClient.setQueriesData(
            { queryKey: query.queryKey },
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
        }
      });

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
    const SSE_ENDPOINT = `${apiURL}/workspaces/${workspace?.slug}/notifications/subscribe?token=${session?.token}`;
    const eventSource = new EventSource(SSE_ENDPOINT);

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
  }, [
    posthog,
    session?.token,
    workspace,
    queryClient,
    handleNotification,
    handleWorkspaceUpdate,
  ]);

  return null;
};
