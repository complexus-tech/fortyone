"use client";

import { useQueryClient } from "@tanstack/react-query";
import { usePostHog } from "posthog-js/react";
import { useEffect, useEffectEvent } from "react";
import { calendarKeys, notificationKeys } from "@/constants/keys";
import { getApiUrl } from "@/lib/api-url";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "@/modules/stories/constants";
import { parseWorkspaceRealtimeEvent } from "./workspace-realtime-contract";

const apiURL = getApiUrl();

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
  const queryClient = useQueryClient();
  const connectionWorkspaceSlug = workspace?.slug;

  const invalidateRealtimeCaches = useEffectEvent((workspaceSlug: string) => {
    void queryClient.invalidateQueries({
      queryKey: notificationKeys.all(workspaceSlug),
    });
    void queryClient.invalidateQueries({
      queryKey: storyKeys.all(workspaceSlug),
    });
    void queryClient.invalidateQueries({
      queryKey: calendarKeys.all(workspaceSlug),
    });
  });

  const handleMessage = useEffectEvent(
    (event: MessageEvent, workspaceSlug: string) => {
      try {
        const realtimeEvent = parseWorkspaceRealtimeEvent(event.data);

        if (realtimeEvent.kind === "calendar-updated") {
          void queryClient.invalidateQueries({
            queryKey: calendarKeys.all(workspaceSlug),
          });
          return;
        }

        if (realtimeEvent.kind === "story-updated") {
          void queryClient.invalidateQueries({
            predicate: (query) => {
              const queryKey = query.queryKey;
              return (
                Array.isArray(queryKey) &&
                queryKey[0] === "stories" &&
                queryKey[1] === workspaceSlug &&
                !queryKey.includes("detail")
              );
            },
          });

          queryClient.setQueryData(
            storyKeys.detail(workspaceSlug, realtimeEvent.storyId),
            (oldData: DetailedStory | undefined) => {
              if (!oldData) return oldData;

              return { ...oldData, ...realtimeEvent.changes };
            },
          );

          if (
            "statusId" in realtimeEvent.changes ||
            "completedAt" in realtimeEvent.changes ||
            "autoSchedulingStatus" in realtimeEvent.changes ||
            "autoSchedulingReason" in realtimeEvent.changes ||
            "autoSchedulingUpdatedAt" in realtimeEvent.changes
          ) {
            void queryClient.invalidateQueries({
              queryKey: calendarKeys.all(workspaceSlug),
            });
          }

          return;
        }

        void queryClient.invalidateQueries({
          queryKey: notificationKeys.all(workspaceSlug),
        });

        if (realtimeEvent.entityType === "story") {
          void queryClient.invalidateQueries({
            queryKey: storyKeys.detail(workspaceSlug, realtimeEvent.entityId),
          });
        }
      } catch (error) {
        posthog.captureException(error);
        invalidateRealtimeCaches(workspaceSlug);
      }
    },
  );

  const handleError = useEffectEvent((error: Event) => {
    posthog.captureException(error);
  });

  useEffect(() => {
    if (!connectionWorkspaceSlug) return;

    const eventSource = new EventSource(
      `${apiURL}/workspaces/${connectionWorkspaceSlug}/notifications/subscribe`,
      { withCredentials: true },
    );
    const cacheWorkspaceSlug = connectionWorkspaceSlug.toLowerCase();
    let shouldReconcileOnOpen = false;

    eventSource.onmessage = (event) => {
      handleMessage(event, cacheWorkspaceSlug);
    };
    eventSource.onerror = (error) => {
      shouldReconcileOnOpen = true;
      handleError(error);
    };
    eventSource.onopen = () => {
      if (!shouldReconcileOnOpen) return;

      shouldReconcileOnOpen = false;
      invalidateRealtimeCaches(cacheWorkspaceSlug);
    };

    return () => {
      eventSource.close();
    };
  }, [connectionWorkspaceSlug]); // eslint-disable-line react-hooks/exhaustive-deps -- Effect Events are stable and must not become reconnection dependencies.

  return null;
};
