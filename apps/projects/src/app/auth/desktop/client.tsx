"use client";

import { useRef, useState } from "react";
import { Box, Button, Text } from "ui";
import { post } from "api-client";
import type { DesktopAuthRequest } from "@/lib/desktop-auth";
import type { ApiResponse } from "@/types";
import { Logo } from "@/components/ui";
import {
  DESKTOP_REDIRECT_URI,
  getDesktopRedirectURL,
} from "@/lib/desktop-auth";

export const DesktopAuthorization = ({
  transaction,
  email,
}: {
  transaction: DesktopAuthRequest;
  email: string;
}) => {
  const pending = useRef(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [opened, setOpened] = useState(false);

  const continueInApp = async () => {
    if (pending.current) return;
    pending.current = true;
    setLoading(true);
    setError(null);
    try {
      const response = await post<ApiResponse<{ code: string; state: string }>>(
        "auth/desktop/authorize",
        {
          ...transaction,
          codeChallengeMethod: "S256",
          redirectUri: transaction.redirectUri ?? DESKTOP_REDIRECT_URI,
        },
      );
      if (!response.data || response.data.state !== transaction.state) {
        throw new Error(
          "The sign-in response was invalid. Return to FortyOne and try again.",
        );
      }
      window.location.assign(
        getDesktopRedirectURL(
          response.data.code,
          response.data.state,
          transaction.redirectUri,
        ),
      );
      setOpened(true);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Unable to complete sign-in. Please try again.",
      );
    } finally {
      pending.current = false;
      setLoading(false);
    }
  };

  return (
    <Box className="w-full max-w-md px-6">
      <Logo asIcon />
      <Text as="h1" className="mt-8 mb-4 text-3xl" fontWeight="semibold">
        Continue in FortyOne
      </Text>
      <Text className="mb-6">Sign in to the desktop app as {email}.</Text>
      {error ? (
        <Text className="mb-4" role="alert">
          {error}
        </Text>
      ) : null}
      {opened ? (
        <Text className="mb-4" role="status">
          FortyOne should open shortly. You can close this tab once you are
          signed in.
        </Text>
      ) : null}
      <Button
        color="invert"
        fullWidth
        loading={loading}
        onClick={() => {
          void continueInApp();
        }}
      >
        {opened ? "Open FortyOne again" : "Continue in the desktop app"}
      </Button>
      <Text className="mt-4" color="muted">
        Only continue if you started signing in from FortyOne on this device.
      </Text>
    </Box>
  );
};
