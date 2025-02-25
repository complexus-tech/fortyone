"use client";

import { Box, Text } from "ui";
import { Logo } from "@/components/ui";
import { CreateAccountForm } from "./components/create-account-form";

export const CreateAccount = () => {
  return (
    <Box className="max-w-sm">
      <Logo asIcon className="relative -left-2 h-10 text-white" />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
        Create your account
      </Text>
      <Text className="mb-6" color="muted">
        Create an account to get started.
      </Text>
      <CreateAccountForm />
    </Box>
  );
};
