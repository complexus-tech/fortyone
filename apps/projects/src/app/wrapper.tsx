"use client";

import { Panel, PanelGroup, PanelResizeHandle } from "react-resizable-panels";
import { Sidebar } from "@/components/shared";

export const Wrapper = ({ children }: { children: React.ReactNode }) => {
  return (
    <PanelGroup direction="horizontal">
      <Panel defaultSize={17} maxSize={20} minSize={15}>
        <Sidebar />
      </Panel>
      <PanelResizeHandle />
      <Panel defaultSize={83}>{children}</Panel>
    </PanelGroup>
  );
};
