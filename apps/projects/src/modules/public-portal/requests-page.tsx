"use client";

import { useState } from "react";
import { Box } from "ui";
import { PublicPortalShell } from "./portal-shell";
import { PublicPortalSidebar } from "./sidebar";
import { PublicFeedbackList } from "./feedback-list";
import type { PublicPortal, PublicPortalViewer } from "./types";

export const PublicPortalRequestsPage = ({
  portal,
  viewer,
}: {
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => {
  const [selectedBoardId, setSelectedBoardId] = useState<string>();

  return (
    <PublicPortalShell activeTab="feedback" portal={portal} viewer={viewer}>
      <Box className="mx-auto grid w-full max-w-[78rem] gap-10 px-4 py-8 md:grid-cols-[minmax(0,1fr)_19rem] md:px-6">
        <Box>
          <PublicFeedbackList boardId={selectedBoardId} portal={portal} />
        </Box>
        <PublicPortalSidebar
          onBoardSelect={setSelectedBoardId}
          portal={portal}
          selectedBoardId={selectedBoardId}
          viewer={viewer}
        />
      </Box>
    </PublicPortalShell>
  );
};
