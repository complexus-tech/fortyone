"use client";
import type { ReactNode } from "react";
import { ResizablePanel } from "ui";
import { Sidebar } from "@/components/shared";

export const MainLayout = ({ children }: { children: ReactNode }) => {
  return (
    <ResizablePanel autoSaveId="main-layout" direction="horizontal">
      <ResizablePanel.Panel defaultSize={18} maxSize={20} minSize={16}>
        <Sidebar />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle className="bg-gray-100/70 dark:bg-dark-100/40" />
      <ResizablePanel.Panel defaultSize={82}>{children}</ResizablePanel.Panel>
    </ResizablePanel>
  );
};
