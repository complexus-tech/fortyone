import type { Metadata } from "next";
import { NotificationSettingsPage } from "@/shell/desktop/notification-settings-page";

export const metadata: Metadata = {
  title: "Settings › Notifications",
};

export default function Page() {
  return <NotificationSettingsPage />;
}
