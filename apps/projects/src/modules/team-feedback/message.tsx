"use client";

import { Box, Text } from "ui";

export const SelectTeamFeedbackMessage = () => (
  <Box className="flex h-dvh items-center justify-center px-6">
    <Box>
      <Text align="center" className="mb-3" fontSize="xl">
        Select feedback
      </Text>
      <Text align="center" color="muted">
        Review customer ideas before planning them into your team&apos;s work.
      </Text>
    </Box>
  </Box>
);
