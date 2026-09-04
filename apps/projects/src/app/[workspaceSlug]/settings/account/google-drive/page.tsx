import type { Metadata } from "next";
import { Suspense } from "react";
import { Box, Skeleton } from "ui";
import { GoogleDriveSettings } from "@/modules/settings/account/google-drive";

export const metadata: Metadata = {
  title: "Settings › Google Drive",
};

export default function GoogleDriveSettingsPage() {
  return (
    <Suspense
      fallback={
        <Box aria-label="Loading Google Drive settings" role="status">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="mt-6 h-48 w-full" />
        </Box>
      }
    >
      <GoogleDriveSettings />
    </Suspense>
  );
}
