import type { Metadata } from "next";
import { PreferencesSettings } from "@/modules/settings/account/preferences";
import { DesktopSettings } from "@/shell/desktop/desktop-settings";

export const metadata: Metadata = {
  title: "Settings › Preferences",
};

export default function Page() {
  return (
    <>
      <PreferencesSettings />
      <DesktopSettings section="behavior" />
    </>
  );
}
