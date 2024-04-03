"use client";
import type { ReactNode } from "react";
import { ResizablePanel } from "ui";
import { BodyContainer } from "../shared/body";

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
    <ResizablePanel.Panel defaultSize={75}>{children}</ResizablePanel.Panel>
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
          <ResizablePanel.Handle />
          <ResizablePanel.Panel defaultSize={25} maxSize={40} minSize={20}>
            <BodyContainer className={className}>{children}</BodyContainer>
          </ResizablePanel.Panel>
        </>
      ) : null}
    </>
  );
};

BoardDividedPanel.MainPanel = MainPanel;
BoardDividedPanel.SideBar = SideBar;
