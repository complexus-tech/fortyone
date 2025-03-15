"use client";

import { Box, Text } from "ui";
import { TerminologyPreferences } from "./components/terminology-preferences";

export const WorkspaceTerminologySettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Terminology Settings
      </Text>
      <TerminologyPreferences />
    </Box>
  );
};
