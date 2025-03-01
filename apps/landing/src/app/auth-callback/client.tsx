"use client";

import { Flex, Text } from "ui";
import { useEffect } from "react";
import { redirect } from "next/navigation";
import nProgress from "nprogress";
import { useSession } from "next-auth/react";
import { Logo, Blur } from "@/components/ui";
import { useAnalytics } from "@/hooks";
import { getRedirectUrl } from "@/utils";
import type { Invitation } from "@/types";

export const ClientPage = ({ invitations }: { invitations: Invitation[] }) => {
  const { data: session } = useSession();
  const { analytics } = useAnalytics();

  useEffect(() => {
    nProgress.done();

    if (session) {
      analytics.identify(session.user!.email!, {
        email: session.user!.email!,
        name: session.user!.name!,
      });
      redirect(getRedirectUrl(session, invitations));
    }
  }, [analytics, session, invitations]);

  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Blur className="absolute left-1/2 right-1/2 z-[10] h-[400px] w-[400px] -translate-x-1/2 bg-warning/[0.07]" />
      <Flex align="center" direction="column" justify="center">
        <Logo asIcon className="mb-1 h-20 animate-pulse text-white" />
        <Text color="muted" fontWeight="medium">
          Setting up your workspace...
        </Text>
      </Flex>
    </Flex>
  );
};
