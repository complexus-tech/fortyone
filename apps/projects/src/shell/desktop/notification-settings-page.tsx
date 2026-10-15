"use client";

import { NotificationsSettings } from "@/modules/settings/account/notifications";
import { getDesktopBridge } from "./bridge";
import { useDesktop } from "./desktop-provider";
import { DesktopSettings } from "./desktop-settings";

export const NotificationSettingsPage = () => {
  const { isDesktop } = useDesktop();
  const bridge = isDesktop ? getDesktopBridge() : null;
  const supported = Boolean(bridge?.getPreferences && bridge.updatePreferences);

  return (
    <NotificationsSettings
      desktopNotifications={
        supported ? (
          <DesktopSettings className="mt-0" section="notifications" />
        ) : undefined
      }
    />
  );
};
