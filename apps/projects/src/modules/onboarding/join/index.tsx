import { Box, Button, Text } from "ui";
import {
  JOIN_ONBOARDING_STEPS,
  OnboardingStepper,
} from "@/components/onboarding/onboarding-stepper";
import { Logo } from "@/components/ui/logo";
import type { Invitation } from "@/modules/invitations/public/types";
import { auth } from "@/auth";
import { JoinForm } from "./components/join-form";

export const JoinWorkspace = async ({
  invitation,
  token,
}: {
  invitation: Invitation;
  token: string;
}) => {
  const session = await auth();
  const { email, workspaceName, role } = invitation;
  const canJoin = session?.user.email === email;

  return (
    <Box className="w-full px-6 md:max-w-xl">
      <Logo asIcon />
      <Text as="h1" className="mt-8 mb-6 text-4xl" fontWeight="semibold">
        Your invitation to {workspaceName}
      </Text>
      {session ? (
        <OnboardingStepper currentStep={0} steps={JOIN_ONBOARDING_STEPS} />
      ) : null}
      <Text className="mb-6" color="muted">
        You&apos;ve been invited to join the team at{" "}
        <span className="font-semibold">{workspaceName}</span> as{" "}
        {role === "admin" ? "an" : "a"}{" "}
        <span className="font-semibold capitalize">{role}.</span>
        {!canJoin ? (
          <>
            {" "}
            Sign in with the email{" "}
            <span className="font-semibold">{email}</span> to continue.
          </>
        ) : null}
      </Text>
      {canJoin ? (
        <JoinForm invitation={invitation} token={token} />
      ) : (
        <Button align="center" color="invert" fullWidth href="/">
          Sign in
        </Button>
      )}
    </Box>
  );
};
