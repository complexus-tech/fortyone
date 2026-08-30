"use client";

import { useState } from "react";
import { Button, Container, Text } from "ui";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { Logo } from "@/components/ui";
import type { Workspace } from "@/types";
import type { Team } from "@/modules/teams/types";
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
    <Container className="max-h-dvh max-w-120 overflow-y-auto md:max-w-xl">
      <Logo asIcon />
      <Text as="h1" className="mt-10 mb-6 text-4xl" fontWeight="semibold">
        Build With Your Team
      </Text>
      <Text className="mb-8" color="muted">
        Great objectives are achieved together. Invite your teammates to
        collaborate and align on your organization&apos;s goals.
      </Text>
      <InviteForm onFormChange={setMembers} />
      <Button
        align="center"
        className="mt-4 md:py-3"
        color="invert"
        disabled={!isValid}
        fullWidth
        loading={isLoading}
        loadingText="Inviting members..."
        onClick={handleContinue}
      >
        Invite members
      </Button>
      <Button
        align="center"
        className="mt-2 md:py-3"
        color="tertiary"
        fullWidth
        href={welcomeUrl}
        size="lg"
        variant="naked"
      >
        I&apos;ll do this later
      </Button>
    </Container>
  );
};
