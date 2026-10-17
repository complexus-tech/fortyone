"use client";

import { useQueryClient } from "@tanstack/react-query";
import { usePostHog } from "posthog-js/react";
import { useEffect, useEffectEvent } from "react";
import {
  calendarKeys,
  chatCapabilityKeys,
  notificationKeys,
  workspaceKeys,
} from "@/constants/keys";
import { getApiUrl } from "@/lib/api-url";
import { useSession } from "@/lib/auth/client";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { reconcileChatQueries } from "@/modules/chat/realtime";
import { applyChatTypingEvent, clearChatTyping } from "@/modules/chat/typing";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "@/modules/stories/constants";
import { deliverDesktopNotification } from "@/shell/desktop/notification";
import { DesktopNotificationBadge } from "@/shell/desktop/desktop-notification-badge";
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
  const { data: session } = useSession();
  const connectionAccountId = session?.user.id;
  const connectionWorkspaceId = workspace?.id;
  const connectionWorkspaceSlug = workspace?.slug;

  const invalidateRealtimeCaches = useEffectEvent((workspaceSlug: string) => {
    void queryClient.invalidateQueries({
      queryKey: workspaceKeys.settings(workspaceSlug),
    });
    void queryClient.invalidateQueries({
      queryKey: chatCapabilityKeys.workspace(workspaceSlug),
    });
    void reconcileChatQueries(queryClient, workspaceSlug);
    void queryClient.invalidateQueries({
      queryKey: notificationKeys.all(workspaceSlug),
    });
    void queryClient.invalidateQueries({
      queryKey: storyKeys.all(workspaceSlug),
    });
    void queryClient.invalidateQueries({
      predicate: ({ queryKey }) =>
        queryKey[0] === "native-link-preview" && queryKey[2] === workspaceSlug,
    });
    void queryClient.invalidateQueries({
      queryKey: calendarKeys.all(workspaceSlug),
    });
  });

  const handleMessage = useEffectEvent(
    (event: MessageEvent, workspaceSlug: string, isCurrent: () => boolean) => {
      try {
        const realtimeEvent = parseWorkspaceRealtimeEvent(event.data);

        if (realtimeEvent.kind === "chat-typing") {
          if (
            realtimeEvent.workspaceId === connectionWorkspaceId &&
            realtimeEvent.userId === connectionAccountId
          ) {
            applyChatTypingEvent(realtimeEvent);
          }
          return;
        }

        if (realtimeEvent.kind === "chat-updated") {
          if (
            realtimeEvent.workspaceId === connectionWorkspaceId &&
            realtimeEvent.userId === connectionAccountId
          ) {
            void reconcileChatQueries(
              queryClient,
              workspaceSlug,
              realtimeEvent.conversationId,
            );
          }
          return;
        }

        if (realtimeEvent.kind === "calendar-updated") {
          void queryClient.invalidateQueries({
            queryKey: calendarKeys.all(workspaceSlug),
          });
          return;
        }

        if (realtimeEvent.kind === "story-updated") {
          const { checklistChanged, ...storyChanges } = realtimeEvent.changes;

          void queryClient.invalidateQueries({
            predicate: ({ queryKey, state }) => {
              if (
                queryKey[0] !== "native-link-preview" ||
                queryKey[2] !== workspaceSlug
              )
                return false;
              const preview = state.data as
                | { resource?: { kind: string; value: { id: string } } }
                | undefined;
              return (
                preview?.resource?.kind === "story" &&
                preview.resource.value.id === realtimeEvent.storyId
              );
            },
          });

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

          if (Object.keys(storyChanges).length > 0) {
            queryClient.setQueryData(
              storyKeys.detail(workspaceSlug, realtimeEvent.storyId),
              (oldData: DetailedStory | undefined) => {
                if (!oldData) return oldData;

                return { ...oldData, ...storyChanges };
              },
            );
          }

          if (checklistChanged) {
            // Includes the checklist and activity queries, plus associations
            // when an item has been converted to a related story.
            void queryClient.invalidateQueries({
              queryKey: storyKeys.detail(workspaceSlug, realtimeEvent.storyId),
            });
          }

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

        if (connectionAccountId && connectionWorkspaceId) {
          void deliverDesktopNotification(
            event.data,
            {
              accountId: connectionAccountId,
              workspaceId: connectionWorkspaceId,
              workspaceSlug,
            },
            isCurrent,
          ).catch((error: unknown) => posthog.captureException(error));
        }

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
    let active = true;

    eventSource.onmessage = (event) => {
      if (active) handleMessage(event, cacheWorkspaceSlug, () => active);
    };
    eventSource.onerror = (error) => {
      shouldReconcileOnOpen = true;
      if (connectionWorkspaceId) clearChatTyping(connectionWorkspaceId);
      handleError(error);
    };
    eventSource.onopen = () => {
      if (!shouldReconcileOnOpen) return;

      shouldReconcileOnOpen = false;
      invalidateRealtimeCaches(cacheWorkspaceSlug);
    };

    return () => {
      active = false;
      eventSource.close();
      if (connectionWorkspaceId) clearChatTyping(connectionWorkspaceId);
    };
  }, [connectionAccountId, connectionWorkspaceId, connectionWorkspaceSlug]); // eslint-disable-line react-hooks/exhaustive-deps -- Effect Events are stable and must not become reconnection dependencies.

  return connectionAccountId &&
    connectionWorkspaceId &&
    connectionWorkspaceSlug ? (
    <DesktopNotificationBadge
      accountId={connectionAccountId}
      workspaceId={connectionWorkspaceId}
      workspaceSlug={connectionWorkspaceSlug.toLowerCase()}
    />
  ) : null;
};
