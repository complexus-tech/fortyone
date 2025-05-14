"use client";

import { useState } from "react";
import { Box, Button, Text } from "ui";
import { Logo } from "@/components/ui";
import { InviteForm } from "./components/invite-form";

type Member = {
  email: string;
};

// Simple email validation regex
const isValidEmail = (email: string) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return email.trim() !== "" && emailRegex.test(email);
};

export const InviteTeam = () => {
  const [members, setMembers] = useState<Member[]>([]);

  // Check if at least one valid email has been entered
  const isValid = members.some((m) => isValidEmail(m.email));

  const handleContinue = () => {
    // Filter out members with empty emails before submitting
    const _validMembers = members.filter((m) => isValidEmail(m.email));
    // Add API call to invite members here
  };

  return (
    <Box className="max-h-dvh max-w-lg overflow-y-auto">
      <Logo asIcon className="relative -left-1 h-10 text-white" />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
        Build With Your Team
      </Text>
      <Text className="mb-8" color="muted">
        Great objectives are achieved together. Invite your teammates to
        collaborate and align on your organization&apos;s goals.
      </Text>
      <InviteForm onFormChange={setMembers} />
      <Button
        align="center"
        className="mt-4 md:h-[2.7rem]"
        disabled={!isValid}
        fullWidth
        onClick={handleContinue}
      >
        Continue
      </Button>
      <Button
        align="center"
        className="mt-2 opacity-80 dark:hover:bg-transparent"
        color="tertiary"
        fullWidth
        href="/onboarding/welcome"
        variant="naked"
      >
        I&apos;ll do this later
      </Button>
    </Box>
  );
};
