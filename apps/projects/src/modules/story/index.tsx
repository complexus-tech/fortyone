"use client";
import { ResizablePanel } from "ui";
import { MainDetails } from "./components/main-details";
import { Options } from "./components/options";
import { DetailedStory } from "./types";

export const StoryPage = ({ story }: { story: DetailedStory }) => {
  return (
    <ResizablePanel autoSaveId="story-details" direction="horizontal">
      <ResizablePanel.Panel defaultSize={72}>
        <MainDetails story={story} />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle />
      <ResizablePanel.Panel defaultSize={28} maxSize={35} minSize={20}>
        <Options story={story} />
      </ResizablePanel.Panel>
    </ResizablePanel>
  );
};
