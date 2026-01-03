"use client";
import { Box, ResizablePanel, Button, Text } from "ui";
import { cn } from "lib";
import type { ReactNode } from "react";
import { ArrowLeft2Icon, StoryMissingIcon } from "icons";
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
  const { isPending, data: story } = useStoryById(storyId);

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
        main: 72,
        side: 28,
        maxSide: 35,
        minSide: 26,
      };
    }
    return {
      key: "story-details",
      main: 72,
      side: 28,
      maxSide: 35,
      minSide: 20,
    };
  };

  return (
    <Box
      className={cn("h-dvh", {
        "h-[85dvh] overflow-y-auto": isDialog,
      })}
    >
      {story ? (
        <>
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
        </>
      ) : (
        <Box className="flex h-screen items-center justify-center">
          <Box className="flex flex-col items-center">
            <StoryMissingIcon className="h-20 w-auto rotate-12" />
            <Text className="mt-10 mb-6" fontSize="3xl">
              404: Item not found
            </Text>
            <Text className="mb-6 max-w-md text-center" color="muted">
              This item might not exist or you do not have access to it.
            </Text>
            <Button
              className="gap-1 pl-2"
              color="tertiary"
              href="/my-work"
              leftIcon={<ArrowLeft2Icon className="h-[1.05rem] w-auto" />}
            >
              Go to my work
            </Button>
          </Box>
        </Box>
      )}
    </Box>
  );
};
