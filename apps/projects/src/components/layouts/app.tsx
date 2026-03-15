"use client";
import { type ReactNode } from "react";
import { Box } from "ui";
import { WorkspaceChatLayout } from "@/components/ui/chat";
import { Sidebar } from "../shared/sidebar/sidebar";

export const ApplicationLayout = ({ children }: { children: ReactNode }) => {
  return (
    <>
      <Box className="md:hidden">{children}</Box>
      <Box className="hidden md:flex">
        <Sidebar />
        <Box className="h-dvh min-w-0 flex-1">
          <WorkspaceChatLayout>{children}</WorkspaceChatLayout>
        </Box>
      </Box>
    </>
  );
};
