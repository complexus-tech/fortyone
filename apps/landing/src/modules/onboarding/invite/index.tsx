"use client";

import { useState } from "react";
import { Box, Button, Text } from "ui";
import { Logo } from "@/components/ui";
import { InviteForm } from "./components/invite-form";

type Member = {
  email: string;
  role: string;
};

export const InviteTeam = () => {
  const [members, setMembers] = useState<Member[]>([]);

  const isValid = members.some((m) => m.email.trim() !== "");

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
        className="mt-2 dark:hover:bg-transparent"
        color="tertiary"
        fullWidth
        href="/onboarding/welcome"
        rounded="lg"
        variant="naked"
      >
        I&apos;ll do this later
      </Button>
    </Box>
  );
};
