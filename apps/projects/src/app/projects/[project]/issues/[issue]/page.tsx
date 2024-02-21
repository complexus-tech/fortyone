"use client";
import { ResizablePanel } from "ui";
import { BodyContainer } from "@/components/layout";
import { Header } from "./components";
import { MainDetails, Options } from "./containers";

export default function Page(): JSX.Element {
  return (
    <>
      <Header />
      <BodyContainer className="overflow-y-hidden">
        <ResizablePanel autoSaveId="issue-details" direction="horizontal">
          <ResizablePanel.Panel defaultSize={72}>
            <MainDetails />
          </ResizablePanel.Panel>
          <ResizablePanel.Handle />
          <ResizablePanel.Panel defaultSize={28} maxSize={35} minSize={25}>
            <Options />
          </ResizablePanel.Panel>
        </ResizablePanel>
      </BodyContainer>
    </>
  );
}
