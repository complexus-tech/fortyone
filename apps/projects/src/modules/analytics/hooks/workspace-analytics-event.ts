import { useCallback } from "react";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { trackWorkspaceAnalyticsEvent } from "../queries/track-workspace-analytics-event";
import type { WorkspaceAnalyticsEventPayload } from "../types";

export const useWorkspaceAnalyticsEvent = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  const trackEvent = useCallback(
    (payload: WorkspaceAnalyticsEventPayload) => {
      if (!session || !workspaceSlug) {
        return;
      }

      void trackWorkspaceAnalyticsEvent(
        { session, workspaceSlug },
        {
          ...payload,
          occurredAt: payload.occurredAt ?? new Date().toISOString(),
        },
      ).catch(() => {
        // Tracking should never block the UI action that generated it.
      });
    },
    [session, workspaceSlug],
  );

  return { trackEvent };
};
