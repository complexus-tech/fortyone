import { Box, Text } from "ui";
import {
  JOIN_ONBOARDING_STEPS,
  OnboardingStepper,
} from "@/components/onboarding/onboarding-stepper";
import { Logo } from "@/components/ui/logo";
import { CreateAccountForm } from "./components/create-account-form";

export const CreateAccount = ({ callbackUrl }: { callbackUrl?: string }) => {
  return (
    <Box className="w-full px-6 md:max-w-xl">
      <Logo asIcon />
      <Text as="h1" className="mt-8 mb-6 text-4xl" fontWeight="semibold">
        Complete your profile
      </Text>
      <OnboardingStepper currentStep={1} steps={JOIN_ONBOARDING_STEPS} />
      <CreateAccountForm callbackUrl={callbackUrl} />
    </Box>
  );
};
