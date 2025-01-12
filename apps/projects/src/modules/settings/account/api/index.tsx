"use client";

import { Box, Flex, Text, Button } from "ui";
import { SectionHeader } from "../../components";

export const ApiSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        API Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button className="bg-primary dark:bg-primary" variant="naked">
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
