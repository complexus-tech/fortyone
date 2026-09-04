"use client";

import { useState } from "react";
import { Box, Button, Flex, Text } from "ui";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import {
  JOIN_ONBOARDING_STEPS,
  OnboardingStepper,
} from "@/components/onboarding/onboarding-stepper";
import { Logo } from "@/components/ui/logo";
import type { Workspace } from "@/types/workspace";
import type { Team } from "@/modules/teams/public/types";
import { inviteOnboardingMembers } from "@/modules/invitations/public/onboarding";
import { withOnboardingCallbackUrl } from "@/modules/onboarding/routing";
import { InviteForm } from "./components/invite-form";

type Member = {
  email: string;
};

const isValidEmail = (email: string) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return email.trim() !== "" && emailRegex.test(email);
};

export const InviteTeam = ({
  activeWorkspace,
  callbackUrl,
  teams,
}: {
  activeWorkspace: Workspace;
  callbackUrl?: string;
  teams: Team[];
}) => {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [members, setMembers] = useState<Member[]>([]);
  const isValid = members.some((m) => isValidEmail(m.email));

  const handleContinue = () => {
    const validEmails: string[] = [];
    for (const member of members) {
      if (isValidEmail(member.email)) {
        validEmails.push(member.email.toLowerCase());
      }
    }
    if (validEmails.length === 0) {
      toast.warning("Invalid email addresses", {
        description: "Please enter valid email addresses",
      });
    }

    setIsLoading(true);
    void inviteOnboardingMembers(
      validEmails,
      teams.map((t) => t.id),
      activeWorkspace.slug,
    )
      .then((res) => {
        if (res.error?.message) {
          toast.error("Failed to invite members", {
            description:
              res.error.message ||
              "But don't worry, you can add them later after you've signed in.",
          });
          return false;
        }

        return true;
      })
      .finally(() => {
        setIsLoading(false);
      })
      .then((shouldContinue) => {
        if (shouldContinue) {
          router.push(
            withOnboardingCallbackUrl("/onboarding/welcome", callbackUrl),
          );
        }
      });
  };

  const welcomeUrl = withOnboardingCallbackUrl(
    "/onboarding/welcome",
    callbackUrl,
  );

  return (
    <Box className="w-full px-6 md:max-w-xl">
      <Logo asIcon />
      <Text as="h1" className="mt-8 mb-6 text-4xl" fontWeight="semibold">
        Build With Your Team
      </Text>
      <OnboardingStepper currentStep={2} steps={JOIN_ONBOARDING_STEPS} />
      <InviteForm onFormChange={setMembers} />
      <Flex align="center" className="mt-5 flex-wrap gap-3" justify="between">
        <Button
          align="center"
          className="shrink-0 px-7 md:px-8"
          color="tertiary"
          disabled={isLoading}
          onClick={() => {
            router.push(welcomeUrl);
          }}
          type="button"
          variant="outline"
        >
          Skip
        </Button>
        <Button
          align="center"
          className="ml-auto shrink-0 px-7 md:px-8"
          color="invert"
          disabled={!isValid || isLoading}
          loading={isLoading}
          loadingText="Inviting…"
          onClick={handleContinue}
        >
          Invite members
        </Button>
      </Flex>
    </Box>
  );
};
