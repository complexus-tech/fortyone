"use client";
import type { ReactNode } from "react";
import { ResizablePanel } from "ui";
import { BodyContainer } from "../layout/body";

export const BoardDividedPanel = ({
  children,
  autoSaveId,
}: {
  children: ReactNode;
  autoSaveId?: string;
}) => {
  return (
    <ResizablePanel autoSaveId={autoSaveId} direction="horizontal">
      {children}
    </ResizablePanel>
  );
};

const MainPanel = ({ children }: { children: ReactNode }) => {
  return (
    <ResizablePanel.Panel defaultSize={70}>{children}</ResizablePanel.Panel>
  );
};

const SideBar = ({
  children,
  isExpanded,
  className,
}: {
  children: ReactNode;
  isExpanded: boolean | null;
  className?: string;
}) => {
  return (
    <>
      {isExpanded ? (
        <>
          <ResizablePanel.Handle className="bg-gray-100/70 dark:bg-dark-100/40" />
          <ResizablePanel.Panel defaultSize={30} maxSize={40} minSize={26}>
            <BodyContainer className={className}>{children}</BodyContainer>
          </ResizablePanel.Panel>
        </>
      ) : null}
    </>
  );
};

BoardDividedPanel.MainPanel = MainPanel;
BoardDividedPanel.SideBar = SideBar;
