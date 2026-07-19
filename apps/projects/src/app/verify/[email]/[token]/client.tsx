"use client";

import { Text, Flex } from "ui";
import { useParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { Logo } from "@/components/ui";
import { getAuthCallbackPath } from "@/utils/callback-url";
import { logIn } from "./actions";

export const EmailVerificationCallback = ({
  callbackUrl,
  isMobileApp,
}: {
  callbackUrl?: string;
  isMobileApp: boolean;
}) => {
  const params = useParams<{ email: string; token: string }>();
  const validatedEmail = decodeURIComponent(params.email);
  const validatedToken = decodeURIComponent(params.token);
  const hasValidated = useRef(false);

  useEffect(() => {
    const validate = async () => {
      if (hasValidated.current) {
        return;
      }
      hasValidated.current = true;
      const res = await logIn(validatedEmail, validatedToken);

      if (res.error) {
        const errorPath = `/?error=${encodeURIComponent(res.error)}`;
        window.location.href = callbackUrl
          ? `${errorPath}&callbackUrl=${encodeURIComponent(callbackUrl)}`
          : errorPath;
        return;
      }

      window.location.href = isMobileApp
        ? "/auth-callback?mobileApp=true"
        : getAuthCallbackPath(callbackUrl);
    };

    void validate();
  }, [callbackUrl, validatedEmail, validatedToken, isMobileApp]);

  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Flex align="center" direction="column" justify="center">
        <Logo asIcon className="mb-1 animate-pulse" />
        <Text color="muted" fontWeight="medium">
          Verifying your secure sign-in link...
        </Text>
      </Flex>
    </Flex>
  );
};
