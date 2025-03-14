"use client";
import { ResizablePanel } from "ui";
import { usePathname } from "next/navigation";
import { MainDetails } from "./components/main-details";
import { Options } from "./components/options";

export const StoryPage = ({ storyId }: { storyId: string }) => {
  const pathname = usePathname();
  const isNotifications = pathname.includes("notifications");
  return (
    <ResizablePanel
      autoSaveId={
        isNotifications ? "story-details-notification" : "story-details"
      }
      direction="horizontal"
    >
      <ResizablePanel.Panel defaultSize={72}>
        <MainDetails storyId={storyId} />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle />
      <ResizablePanel.Panel
        defaultSize={isNotifications ? 25 : 28}
        maxSize={isNotifications ? 28 : 35}
        minSize={isNotifications ? 24 : 25}
      >
        <Options storyId={storyId} />
      </ResizablePanel.Panel>
    </ResizablePanel>
  );
};
