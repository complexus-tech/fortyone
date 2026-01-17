"use client";

import { Text, Flex } from "ui";
import { redirect, useParams, useSearchParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { Logo } from "@/components/ui";
import { getRedirectUrl } from "@/utils";
import { getMyInvitations } from "@/lib/queries/get-invitations";
import { useAnalytics } from "@/hooks";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { logIn, getSession } from "./actions";
import { getAuthCode } from "@/lib/queries/get-auth-code";

export const EmailVerificationCallback = () => {
  const params = useParams<{ email: string; token: string }>();
  const searchParams = useSearchParams();
  const isMobile = searchParams?.get("mobile") === "true";
  const validatedEmail = decodeURIComponent(params?.email || "");
  const validatedToken = decodeURIComponent(params?.token || "");
  const { analytics } = useAnalytics();
  const hasValidated = useRef(false);

  useEffect(() => {
    const validate = async () => {
      if (hasValidated.current) {
        return;
      }
      hasValidated.current = true;
      const res = await logIn(validatedEmail, validatedToken);
      if (res?.error) {
        const session = await getSession();
        if (session) {
          const [workspaces, profile] = await Promise.all([
            getWorkspaces(session?.token || ""),
            getProfile(session!),
          ]);
          if (isMobile && workspaces.length > 0) {
            const authCodeResponse = await getAuthCode(session);
            if (authCodeResponse.error || !authCodeResponse.data) {
              redirect("/login?mobile=true&error=Failed to generate auth code");
              return;
            }
            redirect(
              `fortyone://login?code=${authCodeResponse.data.code}&email=${authCodeResponse.data.email}`,
            );
            return;
          }
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
          return;
        }
        redirect(`/login?error=${res.error}`);
      } else {
        const session = await getSession();
        const [workspaces, profile] = await Promise.all([
          getWorkspaces(session?.token || ""),
          getProfile(session!),
        ]);
        if (session) {
          if (isMobile && workspaces.length > 0) {
            const authCodeResponse = await getAuthCode(session);
            if (authCodeResponse.error || !authCodeResponse.data) {
              redirect("/login?mobile=true&error=Failed to generate auth code");
            } else {
              redirect(
                `fortyone://login?code=${authCodeResponse.data.code}&email=${authCodeResponse.data.email}`,
              );
            }
          }
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
  }, [validatedEmail, validatedToken, analytics, isMobile]);

  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Flex align="center" direction="column" justify="center">
        <Logo asIcon className="mb-1 animate-pulse" />
        <Text color="muted" fontWeight="medium">
          Verifying your secure sign-in link...
        </Text>
      </Flex>
    </Flex>
  );
};
