import type { Metadata } from "next";
import { SelectNotificationMessage } from "@/modules/notifications/message";

export const metadata: Metadata = {
  title: "Notifications",
};

export default function Page() {
  return <SelectNotificationMessage />;
}
