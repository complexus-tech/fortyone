/* eslint-disable @typescript-eslint/no-misused-promises -- TODO: fix */
"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Flex, Input, Text, Button } from "ui";
import { toast } from "sonner";
import nProgress from "nprogress";
import type { Session } from "next-auth";
import { redirect } from "next/navigation";
import { Logo, Blur } from "@/components/ui";
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
        redirect(getRedirectUrl(session));
      }
    } finally {
      setLoading(false);
      nProgress.done();
    }
  };

  return (
    <Flex align="center" className="relative h-screen" justify="center">
      <Blur className="absolute -top-96 left-1/2 right-1/2 h-[300px] w-[300px] -translate-x-1/2 bg-warning/5 md:h-[700px] md:w-[90vw]" />
      <Box className="mx-auto w-full max-w-sm rounded-xl">
        <Logo asIcon className="mx-auto text-white" />
        <Text
          align="center"
          as="h1"
          className="mb-10 mt-6"
          fontSize="3xl"
          fontWeight="medium"
        >
          Log in to your account
        </Text>
        <form onSubmit={handleSubmit}>
          <Input
            autoFocus
            className="mb-4 h-12 rounded-lg"
            label="Email"
            name="email"
            placeholder="e.g john@example.com"
            required
            type="email"
          />
          <Input
            className="mb-5 h-12 rounded-lg"
            label="Password"
            name="password"
            required
            type="password"
          />
          <Button
            align="center"
            className="md:h-12"
            fullWidth
            loading={loading}
            loadingText="Logging in..."
            type="submit"
          >
            Login
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
