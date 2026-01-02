"use client";
import { type ReactNode } from "react";
import { Box, ResizablePanel } from "ui";
import { Sidebar } from "../shared/sidebar/sidebar";

export const ApplicationLayout = ({ children }: { children: ReactNode }) => {
  // const [isPanelCollapsed, setIsPanelCollapsed] = useState(false);

  // const handleLayout = (sizes: number[]) => {
  //   // Check if the first panel (index 0) is collapsed
  //   // You can adjust the threshold as needed (e.g., < 5%)
  //   const isCollapsed = sizes[0] < 12;
  //   setIsPanelCollapsed(isCollapsed);
  // };

  return (
    <>
      <Box className="md:hidden">{children}</Box>
      <Box className="hidden md:block">
        <ResizablePanel
          autoSaveId="application:layout"
          direction="horizontal"
          // onLayout={handleLayout}
        >
          <ResizablePanel.Panel
            // collapsedSize={5}
            // collapsible
            defaultSize={14}
            maxSize={20}
            minSize={12}
          >
            <Sidebar />
          </ResizablePanel.Panel>
          <ResizablePanel.Handle className="bg-border-strong z-2" />
          <ResizablePanel.Panel defaultSize={85}>
            {children}
          </ResizablePanel.Panel>
        </ResizablePanel>
      </Box>
    </>
  );
};
