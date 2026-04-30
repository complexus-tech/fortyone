"use client";

import { Box, Text } from "ui";

export const SelectIntegrationRequestMessage = () => (
  <Box className="flex h-dvh items-center justify-center px-6">
    <Box>
      <Text align="center" className="mb-3" fontSize="xl">
        Select a request
      </Text>
      <Text align="center" color="muted">
        Review incoming integration work before accepting it into the team.
      </Text>
    </Box>
  </Box>
);
