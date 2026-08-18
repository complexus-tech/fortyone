"use client";
import { type ReactNode } from "react";
import { Box } from "ui";
import { WorkspaceChatLayout } from "@/components/ui/chat";
import { SidebarProvider } from "../shared/sidebar/sidebar-context";
import { Sidebar } from "../shared/sidebar/sidebar";

export const ApplicationLayout = ({ children }: { children: ReactNode }) => {
  return (
    <SidebarProvider>
      <Box className="md:flex">
        <Box className="hidden md:block">
          <Sidebar />
        </Box>
        <Box className="h-dvh min-w-0 flex-1">
          <WorkspaceChatLayout>{children}</WorkspaceChatLayout>
        </Box>
      </Box>
    </SidebarProvider>
  );
};
