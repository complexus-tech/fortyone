"use client";
import { Box, ResizablePanel } from "ui";
import { MainDetailsSkeleton } from "./main-details-skeleton";
import { OptionsSkeleton } from "./options-skeleton";

export const StorySkeleton = ({
  isNotifications,
}: {
  isNotifications?: boolean;
}) => {
  return (
    <Box>
      <Box className="md:hidden">
        <MainDetailsSkeleton />
      </Box>
      <Box className="hidden md:block">
        <ResizablePanel
          autoSaveId={
            isNotifications ? "story-details-notification" : "story-details"
          }
          direction="horizontal"
        >
          <ResizablePanel.Panel defaultSize={72}>
            <MainDetailsSkeleton />
          </ResizablePanel.Panel>
          <ResizablePanel.Handle />
          <ResizablePanel.Panel
            defaultSize={isNotifications ? 25 : 28}
            maxSize={isNotifications ? 28 : 35}
            minSize={isNotifications ? 24 : 25}
          >
            <OptionsSkeleton />
          </ResizablePanel.Panel>
        </ResizablePanel>
      </Box>
    </Box>
  );
};
