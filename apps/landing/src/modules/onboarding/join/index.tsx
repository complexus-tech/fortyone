"use client";

import { Box, Text } from "ui";
import { Logo } from "@/components/ui";
import { CreateAccountForm } from "./components/create-account-form";

export const JoinWorkspace = () => {
  return (
    <Box className="w-full max-w-md">
      <Logo asIcon className="relative -left-2 h-10 text-white" />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
        Join your team&apos;s workspace
      </Text>
      <Text className="mb-6" color="muted">
        Connect with your team and start collaborating.
      </Text>
      <CreateAccountForm />
    </Box>
  );
};
