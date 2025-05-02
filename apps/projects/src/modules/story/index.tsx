"use client";
import { Box, ResizablePanel } from "ui";
import { MainDetails } from "./components/main-details";
import { Options } from "./components/options";
import { useStoryById } from "./hooks/story";
import { StorySkeleton } from "./components/story-skeleton";

export const StoryPage = ({
  storyId,
  isNotifications,
}: {
  storyId: string;
  isNotifications?: boolean;
}) => {
  const { isPending } = useStoryById(storyId);

  if (isPending) {
    return <StorySkeleton isNotifications={isNotifications} />;
  }

  return (
    <Box className="h-dvh">
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
          <ResizablePanel.Panel defaultSize={isNotifications ? 75 : 72}>
            <MainDetails
              isNotifications={Boolean(isNotifications)}
              storyId={storyId}
            />
          </ResizablePanel.Panel>
          <ResizablePanel.Handle />
          <ResizablePanel.Panel
            defaultSize={isNotifications ? 25 : 28}
            maxSize={isNotifications ? 28 : 35}
            minSize={isNotifications ? 22 : 25}
          >
            <Options
              isNotifications={Boolean(isNotifications)}
              storyId={storyId}
            />
          </ResizablePanel.Panel>
        </ResizablePanel>
      </Box>
    </Box>
  );
};
