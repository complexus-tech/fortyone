"use client";
import { Box } from "ui";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Projects } from "./projects";

export const Sidebar = () => {
  return (
    <Box className="h-screen border-r border-gray-100 px-4 dark:border-dark-100">
      <Header />
      <Navigation />
      <Projects />
    </Box>
  );
};
