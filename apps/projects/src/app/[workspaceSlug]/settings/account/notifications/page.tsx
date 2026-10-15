import type { Metadata } from "next";
import { NotificationsSettings } from "@/modules/settings/account/notifications";
import { DesktopSettings } from "@/shell/desktop/desktop-settings";

export const metadata: Metadata = {
  title: "Settings › Notifications",
};

export default function Page() {
  return (
    <>
      <NotificationsSettings />
      <DesktopSettings section="notifications" />
    </>
  );
}
