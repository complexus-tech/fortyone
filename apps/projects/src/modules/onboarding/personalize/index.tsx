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
          <Text align="center" as="h1" className="mb-6" fontSize="4xl">
            Tailor Your Workspace
          </Text>
          <Text align="center" color="muted">
            Select your primary domain to help us tailor your experience. This
            will help optimize how you track and achieve your team&apos;s
            objectives.
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
