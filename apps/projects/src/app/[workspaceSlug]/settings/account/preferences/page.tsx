import type { Metadata } from "next";
import { PreferencesSettingsPage } from "@/shell/desktop/preferences-settings-page";

export const metadata: Metadata = {
  title: "Settings › Preferences",
};

export default function Page() {
  return <PreferencesSettingsPage />;
}
