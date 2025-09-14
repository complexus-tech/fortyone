import type { Metadata } from "next";
import { PreferencesSettings } from "@/modules/settings/account/preferences";

export const metadata: Metadata = {
  title: "Settings › Preferences",
};

export default function Page() {
  return <PreferencesSettings />;
}
