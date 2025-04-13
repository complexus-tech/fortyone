"use client";
import type { ReactNode } from "react";
import { ResizablePanel } from "ui";
import { useMediaQuery } from "@/hooks";
import { Sidebar } from "../shared/sidebar/sidebar";

export const ApplicationLayout = ({ children }: { children: ReactNode }) => {
  const isMobile = useMediaQuery("(max-width: 768px)");

  if (isMobile) {
    return <>{children}</>;
  }

  return (
    <ResizablePanel autoSaveId="application:layout" direction="horizontal">
      <ResizablePanel.Panel defaultSize={15} maxSize={20} minSize={12}>
        <Sidebar />
      </ResizablePanel.Panel>
      <ResizablePanel.Handle className="z-[2] -mr-px" />
      <ResizablePanel.Panel defaultSize={85}>{children}</ResizablePanel.Panel>
    </ResizablePanel>
  );
};
