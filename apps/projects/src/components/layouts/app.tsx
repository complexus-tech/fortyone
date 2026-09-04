"use client";
import { type ReactNode } from "react";
import { Box } from "ui";
import { WorkspaceChatLayout } from "@/components/ui/chat";
import { WalkthroughIntegration } from "@/components/walkthrough/walkthrough-integration";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { AppCommandBar } from "@/shell/commands/app-command-bar";
import { AppCommandActionProvider } from "../shared/app-command-action-context";
import { SidebarProvider } from "../shared/sidebar/sidebar-context";
import { Sidebar } from "../shared/sidebar/sidebar";
import { SidebarEdgeToggle } from "../shared/sidebar/sidebar-edge-toggle";

export const ApplicationLayout = ({ children }: { children: ReactNode }) => {
  return (
    <SidebarProvider>
      <AppCommandActionProvider>
        <Box className="flex h-dvh flex-col" data-app-shell>
          <AppCommandBar />
          <Box className="min-h-0 flex-1 md:flex">
            <Box className="relative hidden md:block">
              <Sidebar />
              <SidebarEdgeToggle />
            </Box>
            <Box className="h-dvh min-w-0 flex-1 md:h-full md:pr-(--app-content-inset) md:pb-(--app-content-inset) md:pl-2">
              <Box
                className="app-content-canvas-gradient border-border/80 bg-surface-muted/60 dark:bg-surface-muted/50 h-full min-w-0 overflow-hidden md:rounded-2xl md:border-[0.5px]"
                data-app-content-canvas
                data-walkthrough-target={walkthroughTargets.workspaceContent}
              >
                <WorkspaceChatLayout>{children}</WorkspaceChatLayout>
              </Box>
            </Box>
          </Box>
        </Box>
        <WalkthroughIntegration />
      </AppCommandActionProvider>
    </SidebarProvider>
  );
};
