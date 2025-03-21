import type { Metadata } from "next";
import { Box } from "ui";
import type { ReactNode } from "react";
import { BodyContainer } from "@/components/shared";
import { NotificationsContainer } from "@/modules/notifications/container";

export const metadata: Metadata = {
  title: "Notifications",
};

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <BodyContainer className="h-screen">
      <Box className="grid grid-cols-[320px_auto]">
        <NotificationsContainer />
        {children}
      </Box>
    </BodyContainer>
  );
}
