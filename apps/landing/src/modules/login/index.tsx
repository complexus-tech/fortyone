/* eslint-disable @typescript-eslint/no-misused-promises -- TODO: fix */
"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Input, Text, Button, Flex } from "ui";
import { toast } from "sonner";
import nProgress from "nprogress";
import type { Session } from "next-auth";
import { redirect } from "next/navigation";
import Link from "next/link";
import { Logo, GoogleIcon } from "@/components/ui";
import { useAnalytics } from "@/hooks";
import { getSession, logIn } from "./actions";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

const getRedirectUrl = (session: Session | null) => {
  if (!session) {
    return "/";
  }

  if (session.workspaces.length === 0) {
    return "/onboarding/create";
  }
  const activeWorkspace = session.activeWorkspace || session.workspaces[0];

  if (domain.includes("localhost")) {
    return `http://${activeWorkspace.slug}.localhost:3000/my-work`;
  }

  return `https://${activeWorkspace.slug}.${domain}/my-work`;
};

export const LoginPage = () => {
  const { analytics } = useAnalytics();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    setLoading(true);
    nProgress.start();

    const result = await logIn(formData);
    if (result?.error) {
      toast.error("Failed to log in", {
        description: result.error,
      });
      setLoading(false);
      nProgress.done();
    } else {
      const session = await getSession();
      if (session) {
        analytics.identify(session.user!.email!, {
          email: session.user!.email!,
          name: session.user!.name!,
        });
      }
      redirect(getRedirectUrl(session));
    }
  };

  return (
    <Flex align="center" className="z-[3] bg-[#000000]" justify="center">
      <Box className="max-w-sm">
        <Logo asIcon className="relative -left-2 h-10 text-white" />
        <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
          Sign in to Complexus
        </Text>
        <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
          Don&apos;t have an account?{" "}
          <Link className="text-primary" href="/signup">
            Create one
          </Link>
        </Text>
        <form onSubmit={handleSubmit}>
          <Input
            autoFocus
            className="mb-4 rounded-lg"
            label="Enter your email"
            name="email"
            placeholder="e.g john@example.com"
            required
            type="email"
          />
          <Input
            className="mb-5 rounded-lg"
            label="Password"
            name="password"
            required
            type="password"
          />
          <Button
            align="center"
            fullWidth
            loading={loading}
            loadingText="Logging in..."
            type="submit"
          >
            Continue
          </Button>
          <Flex align="center" className="my-4 gap-4" justify="between">
            <Box className="h-px w-full bg-white/20" />
            <Text className="opacity-40" color="white">
              OR
            </Text>
            <Box className="h-px w-full bg-white/20" />
          </Flex>
          <Button
            align="center"
            className="mb-3"
            color="tertiary"
            fullWidth
            leftIcon={<GoogleIcon />}
            type="button"
          >
            Continue with Google
          </Button>
        </form>
        <Text className="mt-3 pl-[1px] text-[90%]" color="muted">
          &copy; {new Date().getFullYear()} &bull; Powered by Complexus &bull;
          All Rights Reserved.
        </Text>
      </Box>
    </Flex>
  );
};
