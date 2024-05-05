"use client";
import { ResizablePanel } from "ui";
import { MainDetails, Options } from "./containers";

export default function Page(): JSX.Element {
  return (
    <ResizablePanel autoSaveId="story-details" direction="horizontal">
      <ResizablePanel.Panel defaultSize={72}>
        <MainDetails />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle />
      <ResizablePanel.Panel defaultSize={28} maxSize={35} minSize={20}>
        <Options />
      </ResizablePanel.Panel>
    </ResizablePanel>
  );
}
