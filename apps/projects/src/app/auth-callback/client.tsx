"use client";

import { Flex, Text } from "ui";
import { useEffect } from "react";
import nProgress from "nprogress";
import { Logo } from "@/components/ui";
import type { Session } from "@/auth";
import { useAnalytics } from "@/hooks";
import { getRedirectUrl } from "@/utils";
import type { Invitation } from "@/modules/invitations/types";
import type { User, Workspace } from "@/types";

export const ClientPage = ({
  invitations,
  callbackUrl,
  session,
  workspaces,
  profile,
}: {
  invitations: Invitation[];
  callbackUrl?: string;
  session: Session;
  workspaces: Workspace[];
  profile: User;
}) => {
  const { analytics } = useAnalytics();

  useEffect(() => {
    nProgress.done();
    analytics.identify(session.user.email, {
      email: session.user.email,
      name: session.user.name,
    });
    window.location.href = getRedirectUrl(
      workspaces,
      invitations,
      profile.lastUsedWorkspaceId,
      callbackUrl,
    );
  }, [analytics, callbackUrl, session, invitations, workspaces, profile]);

  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Flex align="center" direction="column" justify="center">
        <Logo asIcon className="mb-1 animate-pulse" />
        <Text color="muted" fontWeight="medium">
          Signing into your workspace...
        </Text>
      </Flex>
    </Flex>
  );
};
