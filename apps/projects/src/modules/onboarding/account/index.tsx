import { Box, Text } from "ui";
import { Logo } from "@/components/ui";
import { CreateAccountForm } from "./components/create-account-form";

export const CreateAccount = () => {
  return (
    <Box className="w-full px-6 md:max-w-xl">
      <Logo asIcon />
      <Text as="h1" className="mt-10 mb-6 text-4xl" fontWeight="semibold">
        Create your account
      </Text>
      <Text className="mb-6" color="muted">
        Create an account to get started.
      </Text>
      <CreateAccountForm />
    </Box>
  );
};
