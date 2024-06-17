"use client";
import { ResizablePanel } from "ui";
import { MainDetails } from "./details";
import { Options } from "./options";
import { DetailedStory } from "@/modules/teams/story/types";

export const MainStory = ({ story }: { story: DetailedStory }) => {
  return (
    <ResizablePanel autoSaveId="story-details" direction="horizontal">
      <ResizablePanel.Panel defaultSize={72}>
        <MainDetails story={story} />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle />
      <ResizablePanel.Panel defaultSize={28} maxSize={35} minSize={20}>
        <Options />
      </ResizablePanel.Panel>
    </ResizablePanel>
  );
};
