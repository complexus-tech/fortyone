/* eslint-disable @typescript-eslint/no-misused-promises -- TODO: fix */
"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Input, Text, Button, Flex, BlurImage, Container } from "ui";
import { toast } from "sonner";
import nProgress from "nprogress";
import type { Session } from "next-auth";
import { redirect } from "next/navigation";
import Link from "next/link";
import { FacebookIcon, InstagramIcon, LinkedinIcon, TwitterIcon } from "icons";
import { Logo, Blur } from "@/components/ui";
import { useAnalytics } from "@/hooks";
import { getSession, logIn } from "./actions";
import { GoogleIcon } from "./components/google-icon";
import { AppleIcon } from "./components/apple-icon";

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
    try {
      const result = await logIn(formData);
      if (result?.error) {
        toast.error("Failed to log in", {
          description: result.error,
        });
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
    } finally {
      setLoading(false);
      nProgress.done();
    }
  };

  return (
    <Box className="relative grid h-screen grid-cols-2">
      <Blur className="absolute -top-96 left-1/2 right-1/2 z-10 h-[300px] w-[300px] -translate-x-1/2 bg-warning/5 md:h-[700px] md:w-[90vw]" />
      <Flex align="center" className="z-20 bg-[#000000]" justify="center">
        <Box>
          <Logo asIcon className="relative -left-2 h-10 text-white" />
          <Text
            as="h1"
            className="mb-2 mt-6"
            fontSize="3xl"
            fontWeight="semibold"
          >
            Welcome to Complexus 👋
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
            <Button
              align="center"
              color="tertiary"
              fullWidth
              leftIcon={<AppleIcon />}
              type="button"
            >
              Continue with Apple
            </Button>
          </form>
          <Text className="mt-3 pl-[1px] text-[90%]" color="muted">
            &copy; {new Date().getFullYear()} &bull; Powered by Complexus &bull;
            All Rights Reserved.
          </Text>
        </Box>
      </Flex>
      <Box className="relative bg-black">
        <BlurImage
          alt="Login"
          className="h-full w-full object-cover opacity-80"
          quality={100}
          src="/images/login.webp"
        />
        <Container className="absolute inset-0 z-10 flex flex-col items-center justify-end py-40">
          <Text align="center" className="mb-8 opacity-80" color="muted">
            Connect with us
          </Text>
          <Box className="3xl:gap-16 flex gap-8">
            <Link href="https://x.com/complexus_app" target="_blank">
              <TwitterIcon className="dark:hover:text-primary" />
            </Link>
            <Link
              href="https://www.linkedin.com/company/complexus-tech/"
              target="_blank"
            >
              <LinkedinIcon className="dark:hover:text-primary" />
            </Link>
            <Link
              href="https://www.instagram.com/complexus_tech/"
              target="_blank"
            >
              <InstagramIcon className="dark:hover:text-primary" />
            </Link>
            <Link
              href="https://www.facebook.com/complexus.tech"
              target="_blank"
            >
              <FacebookIcon className="dark:hover:text-primary" />
            </Link>
          </Box>
        </Container>
      </Box>
    </Box>
  );
};
