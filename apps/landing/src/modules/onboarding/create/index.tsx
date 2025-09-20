"use client";

import { Box, Text } from "ui";
import { Logo } from "@/components/ui";
import { CreateWorkspaceForm } from "./components/create-workspace-form";

export const CreateWorkspace = () => {
  return (
    <Box className="px-6 md:max-w-sm md:px-0">
      <Logo />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
        Create Workspace
      </Text>
      <Text className="mb-6" color="muted">
        Create an environment where teams can set objectives, track OKRs, and
        drive results together.
      </Text>
      <CreateWorkspaceForm />
    </Box>
  );
};
