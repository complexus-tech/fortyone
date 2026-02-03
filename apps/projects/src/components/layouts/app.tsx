"use client";
import { type ReactNode } from "react";
import { Box } from "ui";
import { Sidebar } from "../shared/sidebar/sidebar";

export const ApplicationLayout = ({ children }: { children: ReactNode }) => {
  return (
    <>
      <Box className="md:hidden">{children}</Box>
      <Box className="hidden md:flex">
        <Sidebar />
        <Box className="w-[calc(100vw-var(--sidebar-width))] flex-1">
          {children}
        </Box>
      </Box>
    </>
  );
};
