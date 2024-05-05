"use client";
import type { ReactNode } from "react";
import { ResizablePanel } from "ui";
import { Sidebar } from "../shared/sidebar/sidebar";

export const ApplicationLayout = ({ children }: { children: ReactNode }) => {
  return (
    <ResizablePanel autoSaveId="application:layout" direction="horizontal">
      <ResizablePanel.Panel defaultSize={18} maxSize={20} minSize={14}>
        <Sidebar />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle className="mr-[0.5px]" />
      <ResizablePanel.Panel defaultSize={82}>{children}</ResizablePanel.Panel>
    </ResizablePanel>
  );
};
