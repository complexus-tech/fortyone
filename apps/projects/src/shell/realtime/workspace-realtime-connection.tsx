"use client";

import { useQueryClient } from "@tanstack/react-query";
import { usePostHog } from "posthog-js/react";
import { useEffect, useEffectEvent } from "react";
import { calendarKeys, notificationKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getApiUrl } from "@/lib/api-url";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import type { AppNotification } from "@/modules/notifications/types";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "@/modules/stories/constants";
import type { AutoSchedulingStatus, Story } from "@/modules/stories/types";

const apiURL = getApiUrl();

type WorkspaceUpdate = {
  type: "story.workspace_update";
  storyId: string;
  workspaceId: string;
  changes: {
    statusId?: string;
    completedAt?: string | null;
    assigneeId?: string;
    priority?: string;
    title?: string;
    autoSchedulingEnabled?: boolean;
    autoSchedulingLocked?: boolean;
    autoSchedulingStatus?: AutoSchedulingStatus;
    autoSchedulingReason?: string | null;
    autoSchedulingUpdatedAt?: string | null;
  };
  actorId: string;
  actorName: string;
  timestamp: number;
};

/**
 * Keeps the one workspace-wide SSE transport alive for the active workspace.
 *
 * This shell owns only transport lifecycle and dispatch. Resource-specific
 * event validation and cache reducers remain candidates for feature ownership
 * as those modules are migrated.
 */
export const WorkspaceRealtimeConnection = () => {
  const posthog = usePostHog();
  const { workspace } = useCurrentWorkspace();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const connectionWorkspaceSlug = workspace?.slug;

  const handleMessage = useEffectEvent((event: MessageEvent) => {
    try {
      const data = JSON.parse(`${event.data}`);

      if (data.type === "calendar.updated") {
        void queryClient.invalidateQueries({
          queryKey: calendarKeys.all(workspaceSlug),
        });
        return;
      }

      if (data.type === "story.workspace_update") {
        const workspaceUpdate = data as WorkspaceUpdate;

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

            return oldData.map((story) =>
              story.id === workspaceUpdate.storyId
                ? { ...story, ...workspaceUpdate.changes }
                : story,
            );
          },
        );

        queryClient.setQueryData(
          storyKeys.detail(workspaceSlug, workspaceUpdate.storyId),
          (oldData: DetailedStory | undefined) => {
            if (!oldData) return oldData;

            return { ...oldData, ...workspaceUpdate.changes };
          },
        );

        if (
          "statusId" in workspaceUpdate.changes ||
          "completedAt" in workspaceUpdate.changes ||
          "autoSchedulingStatus" in workspaceUpdate.changes ||
          "autoSchedulingReason" in workspaceUpdate.changes ||
          "autoSchedulingUpdatedAt" in workspaceUpdate.changes
        ) {
          void queryClient.invalidateQueries({
            queryKey: calendarKeys.all(workspaceSlug),
          });
        }

        return;
      }

      const notification = data as AppNotification;
      void queryClient.invalidateQueries({
        queryKey: notificationKeys.all(workspaceSlug),
      });

      if (notification.entityType === "story") {
        void queryClient.invalidateQueries({
          queryKey: storyKeys.detail(workspaceSlug, notification.entityId),
        });
      }
    } catch (error) {
      posthog.captureException(error);
    }
  });

  const handleError = useEffectEvent((error: Event) => {
    posthog.captureException(error);
  });

  useEffect(() => {
    if (!connectionWorkspaceSlug) return;

    const eventSource = new EventSource(
      `${apiURL}/workspaces/${connectionWorkspaceSlug}/notifications/subscribe`,
      { withCredentials: true },
    );

    eventSource.onmessage = (event) => {
      handleMessage(event);
    };
    eventSource.onerror = (error) => {
      handleError(error);
    };

    return () => {
      eventSource.close();
    };
  }, [connectionWorkspaceSlug, workspaceSlug]); // eslint-disable-line react-hooks/exhaustive-deps -- React 19 Effect Events keep handlers current without reconnecting.

  return null;
};
