"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Input, Text, Button, Flex } from "ui";
import { toast } from "sonner";
import nProgress from "nprogress";
import Link from "next/link";
import { Logo, GoogleIcon } from "@/components/ui";
import { useAnalytics } from "@/hooks";
import { getSession, logIn } from "./actions";

export const SignupPage = () => {
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
      }
    } finally {
      setLoading(false);
      nProgress.done();
    }
  };

  return (
    <Box className="max-w-sm">
      <Logo asIcon className="relative -left-2 h-10 text-white" />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
        Get started with Complexus
      </Text>
      <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
        Already have an account?{" "}
        <Link className="text-primary" href="/login">
          Sign in
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
        &copy; {new Date().getFullYear()} &bull; Powered by Complexus &bull; All
        Rights Reserved.
      </Text>
    </Box>
  );
};
