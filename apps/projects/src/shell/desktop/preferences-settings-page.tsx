"use client";

import { PreferencesSettings } from "@/modules/settings/account/preferences";
import { getDesktopBridge } from "./bridge";
import { useDesktop } from "./desktop-provider";
import { DesktopSettings } from "./desktop-settings";

export const PreferencesSettingsPage = () => {
  const { isDesktop } = useDesktop();
  const bridge = isDesktop ? getDesktopBridge() : null;
  const supported = Boolean(bridge?.getPreferences && bridge.updatePreferences);

  return (
    <PreferencesSettings
      desktopSettings={
        supported ? (
          <DesktopSettings className="mt-0" section="behavior" />
        ) : undefined
      }
    />
  );
};
