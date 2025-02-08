"use client";

import { useState } from "react";
import { Box, Button, Flex, Text } from "ui";
import { InviteForm } from "./components/invite-form";

type Member = {
  email: string;
  role: string;
};

export const InviteTeam = () => {
  const [members, setMembers] = useState<Member[]>([]);

  const isValid = members.some((m) => m.email.trim() !== "");

  return (
    <Flex align="center" className="min-h-screen" justify="center">
      <Box className="w-full max-w-xl px-4">
        <Text align="center" as="h1" className="mb-6" fontSize="4xl">
          Build With Your Team
        </Text>
        <Text align="center" className="mb-8" color="muted">
          Great objectives are achieved together. Invite your teammates to
          collaborate and align on your organization&apos;s goals.
        </Text>
        <InviteForm onFormChange={setMembers} />
        <Button
          align="center"
          className="mt-4"
          disabled={!isValid}
          fullWidth
          rounded="lg"
          size="lg"
        >
          Continue
        </Button>
        <Button
          align="center"
          className="mt-2"
          color="tertiary"
          fullWidth
          href="/onboarding/welcome"
          rounded="lg"
          size="lg"
          variant="naked"
        >
          I&apos;ll do this later
        </Button>
      </Box>
    </Flex>
  );
};
