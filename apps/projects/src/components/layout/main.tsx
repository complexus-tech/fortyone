"use client";
import type { ReactNode } from "react";
import { ResizablePanel } from "ui";
import { Sidebar } from "@/components/shared";

export const MainLayout = ({ children }: { children: ReactNode }) => {
  return (
    <ResizablePanel direction="horizontal">
      <ResizablePanel.Panel defaultSize={17} maxSize={20} minSize={15}>
        <Sidebar />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle />
      <ResizablePanel.Panel defaultSize={83}>{children}</ResizablePanel.Panel>
    </ResizablePanel>
  );
};
