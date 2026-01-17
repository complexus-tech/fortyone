import type { Metadata } from "next";
import { NotificationsSettings } from "@/modules/settings/account/notifications";

export const metadata: Metadata = {
  title: "Settings › Notifications",
};

export default function Page() {
  return <NotificationsSettings />;
}
