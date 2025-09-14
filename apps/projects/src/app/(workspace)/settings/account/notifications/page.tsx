import type { Metadata } from "next";
import { NotificationsSettings } from "@/modules/settings/account/notifications";

export const metadata: Metadata = {
  title: "Settings › Account › Notifications",
};

export default function Page() {
  return <NotificationsSettings />;
}
