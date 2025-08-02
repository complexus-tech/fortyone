"use client";

import { Text, Flex } from "ui";
import { redirect, useParams } from "next/navigation";
import { useEffect } from "react";
import { Logo } from "@/components/ui";
import { getRedirectUrl } from "@/utils";
import { getMyInvitations } from "@/lib/queries/get-invitations";
import { useAnalytics } from "@/hooks";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { logIn, getSession } from "./actions";

export const EmailVerificationCallback = () => {
  const params = useParams<{ email: string; token: string }>();
  const validatedEmail = decodeURIComponent(params?.email || "");
  const validatedToken = decodeURIComponent(params?.token || "");
  const { analytics } = useAnalytics();

  useEffect(() => {
    const validate = async () => {
      const res = await logIn(validatedEmail, validatedToken);
      if (res?.error) {
        redirect(`/login?error=${res.error}`);
      } else {
        const session = await getSession();
        const [workspaces, profile] = await Promise.all([
          getWorkspaces(session?.token || ""),
          getProfile(session!),
        ]);
        if (session) {
          analytics.identify(session.user!.email!, {
            email: session.user!.email!,
            name: session.user!.name!,
          });
          const invitations = await getMyInvitations();
          redirect(
            getRedirectUrl(
              workspaces,
              invitations.data || [],
              profile?.lastUsedWorkspaceId,
            ),
          );
        }
      }
    };

    void validate();
  }, [validatedEmail, validatedToken, analytics]);

  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Flex align="center" direction="column" justify="center">
        <Logo asIcon className="mb-1 h-20 animate-pulse text-white" />
        <Text color="muted" fontWeight="medium">
          Verifying your secure sign-in link...
        </Text>
      </Flex>
    </Flex>
  );
};
