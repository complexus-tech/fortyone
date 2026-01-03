"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../../components";

export const ApiSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        API Settings
      </Text>

      <Box className="rounded-2xl border border-border bg-surface">
        <SectionHeader
          action={
            <Button
              className="shrink-0 bg-primary dark:bg-primary"
              disabled
              variant="naked"
            >
              Create API Key
            </Button>
          }
          description="Manage your API keys for accessing the API."
          title="API Keys"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No API keys created yet.</Text>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
