"use client";

import { useEffect } from "react";
import { usePostHog } from "posthog-js/react";
import { useUnreadNotifications } from "@/modules/notifications/hooks/unread";
import type { DesktopNotificationScope } from "./bridge";
import { getDesktopBridge } from "./bridge";
import { useDesktop } from "./desktop-provider";
import { bindDesktopSession } from "./session-binding";

export const DesktopNotificationBadge = (scope: DesktopNotificationScope) => {
  const { isDesktop } = useDesktop();
  return isDesktop && getDesktopBridge()?.setNotificationBadge ? (
    <NotificationBadge {...scope} />
  ) : null;
};

const NotificationBadge = ({
  accountId,
  workspaceId,
  workspaceSlug,
}: DesktopNotificationScope) => {
  const { data: count, isSuccess } = useUnreadNotifications();
  const posthog = usePostHog();
  useEffect(() => {
    const bridge = getDesktopBridge();
    if (
      !bridge?.setNotificationBadge ||
      !isSuccess ||
      typeof count !== "number"
    )
      return;
    let current = true;
    const scope = { accountId, workspaceId, workspaceSlug };
    void bindDesktopSession(scope)
      .then(async (bound) => {
        if (bound && current)
          await bridge.setNotificationBadge?.({ ...scope, count });
      })
      .catch((error: unknown) => {
        if (current) posthog.captureException(error);
      });
    return () => {
      current = false;
    };
  }, [accountId, count, isSuccess, posthog, workspaceId, workspaceSlug]);
  return null;
};
