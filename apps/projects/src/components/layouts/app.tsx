"use client";
import { type ReactNode } from "react";
import { Box } from "ui";
import { WorkspaceChatLayout } from "@/components/ui/chat";
import { AppCommandBar } from "../shared/app-command-bar";
import { AppCommandActionProvider } from "../shared/app-command-action-context";
import { SidebarProvider } from "../shared/sidebar/sidebar-context";
import { Sidebar } from "../shared/sidebar/sidebar";

export const ApplicationLayout = ({ children }: { children: ReactNode }) => {
  return (
    <SidebarProvider>
      <AppCommandActionProvider>
        <Box className="flex h-dvh flex-col" data-app-shell>
          <AppCommandBar />
          <Box className="min-h-0 flex-1 md:flex">
            <Box className="hidden md:block">
              <Sidebar />
            </Box>
            <Box className="h-dvh min-w-0 flex-1 md:h-full md:pr-[10px] md:pb-[10px] md:pl-2">
              <Box
                className="border-border bg-background h-full min-w-0 overflow-hidden md:rounded-xl md:border-[0.5px]"
                data-app-content-canvas
              >
                <WorkspaceChatLayout>{children}</WorkspaceChatLayout>
              </Box>
            </Box>
          </Box>
        </Box>
      </AppCommandActionProvider>
    </SidebarProvider>
  );
};
