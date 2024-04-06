"use client";
import { Box } from "ui";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";

export const Sidebar = () => {
  return (
    <Box className="flex h-screen flex-col justify-between bg-gray-100/50 px-4 pb-4 dark:bg-[#101010]">
      <Box>
        <Header />
        <Navigation />
        <Teams />
      </Box>
    </Box>
  );
};
