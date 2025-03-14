import type { Metadata } from "next";
import { BodyContainer } from "@/components/shared";
import { ListNotifications } from "@/modules/notifications/list";

export const metadata: Metadata = {
  title: "Notifications",
};

export default function Page() {
  return (
    <BodyContainer className="h-screen">
      <ListNotifications />
    </BodyContainer>
  );
}
