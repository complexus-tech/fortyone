"use client";

import { useState } from "react";
import { Box, Button, Flex, Text } from "ui";
import { DomainSelector } from "./components/domain-selector";

export const PersonalizeWorkspace = () => {
  const [domain, setDomain] = useState("");

  return (
    <Box className="min-h-screen">
      <Flex align="center" className="min-h-screen" justify="center">
        <Box className="w-full max-w-xl px-4">
          <Text align="center" as="h1" className="mb-2" fontSize="3xl">
            Looking good! Let&apos;s personalize your workspace.
          </Text>
          <Text align="center" className="mb-8" color="muted">
            What is your domain expertise? Choose one.
          </Text>

          <DomainSelector
            onChange={(domain) => {
              setDomain(domain);
            }}
            value={domain}
          />

          <Button
            align="center"
            className="mt-6"
            disabled={!domain}
            fullWidth
            href="/onboarding/invite"
            rounded="lg"
            size="lg"
          >
            Continue
          </Button>
        </Box>
      </Flex>
    </Box>
  );
};
