"use client";

import { useState } from "react";
import { Button, Container, Text } from "ui";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { Logo } from "@/components/ui";
import type { Workspace, Team } from "@/types";
import { inviteMembers } from "@/lib/actions/invite-members";
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
  teams,
}: {
  activeWorkspace: Workspace;
  teams: Team[];
}) => {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [members, setMembers] = useState<Member[]>([]);
  const isValid = members.some((m) => isValidEmail(m.email));

  const handleContinue = async () => {
    const validEmails = members
      .filter((m) => isValidEmail(m.email))
      .map((m) => m.email.toLowerCase());
    if (validEmails.length === 0) {
      toast.warning("Invalid email addresses", {
        description: "Please enter valid email addresses",
      });
    }

    setIsLoading(true);
    const res = await inviteMembers(
      validEmails,
      teams.map((t) => t.id),
      activeWorkspace.slug,
    );
    if (res?.error?.message) {
      setIsLoading(false);
      toast.error("Failed to invite members", {
        description:
          res.error.message ||
          "But don't worry, you can add them later after you've signed in.",
      });
      return;
    }
    setIsLoading(false);
    router.push("/onboarding/welcome");
  };

  return (
    <Container className="max-h-dvh max-w-[30rem] overflow-y-auto md:max-w-lg">
      <Logo />
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
        color="invert"
        disabled={!isValid}
        fullWidth
        loading={isLoading}
        loadingText="Inviting members..."
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
    </Container>
  );
};
