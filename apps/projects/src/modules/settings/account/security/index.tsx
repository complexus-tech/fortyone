"use client";

import { Box, Flex, Text } from "ui";
import { SectionHeader } from "../../components";

export const SecuritySettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Security Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Your account is protected by industry-standard security measures."
          title="Account Security"
        />

        <Box className="p-6">
          <Flex direction="column" gap={4}>
            <Text color="muted">
              Your account is secured using social authentication, which
              provides built-in security features including multi-factor
              authentication.
            </Text>
            <Text color="muted">
              There&apos;s no need for additional security layers - you&apos;re
              already protected by enterprise-grade security measures.
            </Text>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
