"use client";
import { ResizablePanel } from "ui";
import { BodyContainer } from "@/components/layout";
import { Header, Options } from "./components";
import { MainDetails } from "./containers";

export default function Page(): JSX.Element {
  return (
    <>
      <Header />
      <BodyContainer className="overflow-y-hidden">
        <ResizablePanel direction="horizontal">
          <ResizablePanel.Panel defaultSize={74}>
            <MainDetails />
          </ResizablePanel.Panel>
          <ResizablePanel.Handle />
          <ResizablePanel.Panel defaultSize={26} maxSize={35} minSize={25}>
            <Options />
          </ResizablePanel.Panel>
        </ResizablePanel>
      </BodyContainer>
    </>
  );
}
