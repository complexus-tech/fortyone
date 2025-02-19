"use client";

import { Flex, Text } from "ui";
import { useEffect } from "react";
import { redirect, useRouter } from "next/navigation";
import type { Session } from "next-auth";
import nProgress from "nprogress";
import { useSession } from "next-auth/react";
import { Logo, Blur } from "@/components/ui";
import { useAnalytics } from "@/hooks";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

const getRedirectUrl = (session: Session) => {
  if (session.workspaces.length === 0) {
    return "/onboarding/create";
  }
  const activeWorkspace = session.activeWorkspace || session.workspaces[0];
  if (domain.includes("localhost")) {
    return `http://${activeWorkspace.slug}.localhost:3000/my-work`;
  }
  return `https://${activeWorkspace.slug}.${domain}/my-work`;
};

export const ClientPage = () => {
  const { data: session } = useSession();
  const router = useRouter();
  const { analytics } = useAnalytics();

  useEffect(() => {
    nProgress.done();

    if (session) {
      analytics.identify(session.user!.email!, {
        email: session.user!.email!,
        name: session.user!.name!,
      });
      if (getRedirectUrl(session).includes("onboarding")) {
        router.push(getRedirectUrl(session));
      } else {
        redirect(getRedirectUrl(session));
      }
    }
  }, [analytics, session, router]);

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
