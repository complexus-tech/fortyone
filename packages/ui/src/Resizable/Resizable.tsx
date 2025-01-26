"use client";
import * as ResizablePrimitive from "react-resizable-panels";
import { cn } from "lib";
import { Box } from "../Box/Box";

export const ResizablePanel = ({
  children,
  className,
  ...props
}: ResizablePrimitive.PanelGroupProps) => (
  <ResizablePrimitive.PanelGroup
    className={cn(
      "flex h-full w-full data-[panel-group-direction=vertical]:flex-col",
      className
    )}
    {...props}
  >
    {children}
  </ResizablePrimitive.PanelGroup>
);

const ResizableHandle = ({
  className,
  children,
  ...props
}: ResizablePrimitive.PanelResizeHandleProps) => (
  <ResizablePrimitive.PanelResizeHandle
    className={cn(
      "group relative bg-gray-200/70 dark:bg-dark-100/80 flex w-px items-center justify-center after:absolute after:inset-y-0 after:left-1/2 after:w-1 after:-translate-x-1/2 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-1 data-[panel-group-direction=vertical]:h-px data-[panel-group-direction=vertical]:w-full data-[panel-group-direction=vertical]:after:left-0 data-[panel-group-direction=vertical]:after:h-1 data-[panel-group-direction=vertical]:after:w-full data-[panel-group-direction=vertical]:after:-translate-y-1/2 data-[panel-group-direction=vertical]:after:translate-x-0 [&[data-panel-group-direction=vertical]>div]:rotate-90",
      className
    )}
    {...props}
  >
    {children}
    <Box className="absolute z-10 bottom-0 right-0 left-0 top-0 w-[0.15rem] bg-gray-300/80 opacity-0 transition hover:opacity-100 dark:bg-dark-50" />
  </ResizablePrimitive.PanelResizeHandle>
);

ResizablePanel.Handle = ResizableHandle;
ResizablePanel.Panel = ResizablePrimitive.Panel;
