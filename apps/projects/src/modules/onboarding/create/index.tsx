"use client";

import { Box, Flex, Text } from "ui";
import { CreateWorkspaceForm } from "./components/create-workspace-form";

export const CreateWorkspace = () => {
  return (
    <Flex align="center" className="min-h-screen" justify="center">
      <Box className="w-full max-w-xl px-4">
        <Text
          align="center"
          as="h1"
          className="mb-6"
          fontSize="4xl"
          fontWeight="light"
        >
          Create new workspace
        </Text>
        <Text align="center" className="max-w-l mx-auto mb-6" color="muted">
          Workspaces are shared environments where teams can work on objectives,
          okrs, sprints, and more.
        </Text>
        <CreateWorkspaceForm />
      </Box>
    </Flex>
  );
};
