"use client";

import { useRef, useState } from "react";
import { Box, Button, Text } from "ui";
import { post } from "api-client";
import type { ApiResponse } from "@/types";
import type { MobileAuthRequest } from "@/lib/mobile-auth";
import { Logo } from "@/components/ui";
import { getMobileRedirectURL, MOBILE_REDIRECT_URI } from "@/lib/mobile-auth";

export const MobileAuthorization = ({
  transaction,
  email,
}: {
  transaction: MobileAuthRequest;
  email: string;
}) => {
  const pending = useRef(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const continueInApp = async () => {
    if (pending.current) return;
    pending.current = true;
    setLoading(true);
    setError(null);
    try {
      const response = await post<ApiResponse<{ code: string; state: string }>>(
        "auth/mobile/authorize",
        {
          ...transaction,
          codeChallengeMethod: "S256",
          redirectUri: MOBILE_REDIRECT_URI,
        },
      );
      if (!response.data || response.data.state !== transaction.state) {
        throw new Error(
          "The sign-in response was invalid. Return to the app and try again.",
        );
      }
      window.location.assign(
        getMobileRedirectURL(response.data.code, response.data.state),
      );
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
      <Text className="mb-6">Sign in to the mobile app as {email}.</Text>
      {error ? (
        <Text className="mb-4" role="alert">
          {error}
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
        Continue in the app
      </Button>
      <Text className="mt-4" color="muted" fontSize="sm">
        Only continue if you started signing in from the FortyOne app on this
        device.
      </Text>
    </Box>
  );
};
