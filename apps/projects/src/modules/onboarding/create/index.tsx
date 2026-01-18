"use client";

import { Box, Text } from "ui";
import { Logo } from "@/components/ui";
import { CreateWorkspaceForm } from "./components/create-workspace-form";

export const CreateWorkspace = () => {
  return (
    <Box className="px-6 md:max-w-lg">
      <Logo asIcon />
      <Text as="h1" className="mt-10 mb-6 text-4xl" fontWeight="semibold">
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
