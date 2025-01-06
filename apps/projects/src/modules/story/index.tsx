"use client";
import { ResizablePanel } from "ui";
import { MainDetails } from "./components/main-details";
import { Options } from "./components/options";

export const StoryPage = () => {
  return (
    <ResizablePanel autoSaveId="story-details" direction="horizontal">
      <ResizablePanel.Panel defaultSize={72}>
        <MainDetails />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle />
      <ResizablePanel.Panel defaultSize={28} maxSize={35} minSize={25}>
        <Options />
      </ResizablePanel.Panel>
    </ResizablePanel>
  );
};
