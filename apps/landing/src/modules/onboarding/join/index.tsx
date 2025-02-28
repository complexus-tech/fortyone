"use client";

import { Box, Text } from "ui";
import { Logo } from "@/components/ui";
import type { Invitation } from "@/lib/actions/verify-invitation";
import { CreateAccountForm } from "./components/create-account-form";

export const JoinWorkspace = ({ invitation }: { invitation: Invitation }) => {
  return (
    <Box className="w-full max-w-md">
      <Logo asIcon className="relative -left-2 h-10 text-white" />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
        Join {invitation.workspaceName}
      </Text>
      <Text className="mb-6" color="muted">
        You&apos;ve been invited to join the team at {invitation.workspaceName}.
      </Text>
      <CreateAccountForm invitation={invitation} />
    </Box>
  );
};
