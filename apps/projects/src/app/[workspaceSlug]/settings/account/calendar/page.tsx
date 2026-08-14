import type { Metadata } from "next";
import { Suspense } from "react";
import { Box, Skeleton } from "ui";
import { CalendarIntegrationSettings } from "@/modules/settings/workspace/integrations/calendar";

export const metadata: Metadata = {
  title: "Settings › Calendar",
};

export default function CalendarSettingsPage() {
  return (
    <Suspense
      fallback={
        <Box aria-label="Loading calendar settings" role="status">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="mt-6 h-48 w-full" />
        </Box>
      }
    >
      <CalendarIntegrationSettings />
    </Suspense>
  );
}
