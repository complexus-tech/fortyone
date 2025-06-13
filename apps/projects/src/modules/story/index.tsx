"use client";
import { Box, ResizablePanel } from "ui";
import { cn } from "lib";
import type { ReactNode } from "react";
import { MainDetails } from "./components/main-details";
import { Options } from "./components/options";
import { useStoryById } from "./hooks/story";
import { StorySkeleton } from "./components/story-skeleton";

export const StoryPage = ({
  storyId,
  isNotifications,
  isDialog,
  mainHeader,
}: {
  storyId: string;
  isNotifications?: boolean;
  isDialog?: boolean;
  mainHeader?: ReactNode;
}) => {
  const { isPending } = useStoryById(storyId);

  if (isPending) {
    return <StorySkeleton isNotifications={isNotifications} />;
  }

  const getResizablePanelKeyAndSize = () => {
    if (isNotifications) {
      return {
        key: "story-details-notification",
        main: 75,
        side: 25,
        maxSide: 28,
        minSide: 22,
      };
    }
    if (isDialog) {
      return {
        key: "story-details-dialog",
        main: 70,
        side: 30,
        maxSide: 35,
        minSide: 28,
      };
    }
    return {
      key: "story-details",
      main: 72,
      side: 28,
      maxSide: 35,
      minSide: 25,
    };
  };

  return (
    <Box
      className={cn("h-dvh", {
        "h-[85dvh]": isDialog,
      })}
    >
      <Box className="md:hidden">
        <MainDetails
          isNotifications={Boolean(isNotifications)}
          storyId={storyId}
        />
      </Box>
      <Box className="hidden md:block">
        <ResizablePanel
          autoSaveId={
            isNotifications ? "story-details-notification" : "story-details"
          }
          direction="horizontal"
        >
          <ResizablePanel.Panel
            defaultSize={getResizablePanelKeyAndSize().main}
          >
            <MainDetails
              isDialog={isDialog}
              isNotifications={Boolean(isNotifications)}
              mainHeader={mainHeader}
              storyId={storyId}
            />
          </ResizablePanel.Panel>
          <ResizablePanel.Handle />
          <ResizablePanel.Panel
            defaultSize={getResizablePanelKeyAndSize().side}
            maxSize={getResizablePanelKeyAndSize().maxSide}
            minSize={getResizablePanelKeyAndSize().minSide}
          >
            <Options
              isDialog={isDialog}
              isNotifications={Boolean(isNotifications)}
              storyId={storyId}
            />
          </ResizablePanel.Panel>
        </ResizablePanel>
      </Box>
    </Box>
  );
};
