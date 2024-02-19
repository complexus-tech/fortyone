"use client";
import type { ReactNode } from "react";
import { DndContext } from "@dnd-kit/core";
import { ResizablePanel } from "ui";
import { Sidebar } from "@/components/shared";

export const MainLayout = ({ children }: { children: ReactNode }) => {
  return (
    <DndContext>
      <ResizablePanel autoSaveId="main-layout" direction="horizontal">
        <ResizablePanel.Panel defaultSize={18} maxSize={20} minSize={16}>
          <Sidebar />
        </ResizablePanel.Panel>
        <ResizablePanel.Handle />
        <ResizablePanel.Panel defaultSize={82}>{children}</ResizablePanel.Panel>
      </ResizablePanel>
    </DndContext>
  );
};
