"use client";

import { Box, Flex, Text } from "ui";
import { CreateWorkspaceForm } from "./components/create-workspace-form";

export const CreateWorkspace = () => {
  return (
    <Flex align="center" className="min-h-screen" justify="center">
      <Box className="w-full max-w-xl px-4">
        <Text align="center" as="h1" className="mb-6" fontSize="4xl">
          Create Workspace
        </Text>
        <Text align="center" className="max-w-l mx-auto mb-6" color="muted">
          Create a collaborative environment where teams can set objectives,
          track OKRs, and drive results together.
        </Text>
        <CreateWorkspaceForm />
      </Box>
    </Flex>
  );
};
