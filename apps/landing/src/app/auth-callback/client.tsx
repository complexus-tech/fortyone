"use client";

import { Flex, Text } from "ui";
import { useEffect } from "react";
import { redirect, useSearchParams } from "next/navigation";
import nProgress from "nprogress";
import type { Session } from "next-auth";
import { Logo } from "@/components/ui";
import { useAnalytics } from "@/hooks";
import { getRedirectUrl } from "@/utils";
import type { Invitation, User, Workspace } from "@/types";

export const ClientPage = ({
  invitations,
  session,
  workspaces,
  profile,
}: {
  invitations: Invitation[];
  session: Session;
  workspaces: Workspace[];
  profile: User;
}) => {
  const { analytics } = useAnalytics();

  useEffect(() => {
    nProgress.done();
    if (session) {
      analytics.identify(session.user!.email!, {
        email: session.user!.email!,
        name: session.user!.name!,
      });
      redirect(
        getRedirectUrl(workspaces, invitations, profile.lastUsedWorkspaceId),
      );
    }
  }, [analytics, session, invitations, workspaces, profile]);

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
