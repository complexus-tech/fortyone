"use client";

import { Text, Flex } from "ui";
import { useParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { Logo } from "@/components/ui";
import { getAuthCallbackPath, getLoginUrl } from "@/utils/callback-url";
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
        const errorURL = new URL(
          getLoginUrl(callbackUrl, isMobileApp),
          window.location.origin,
        );
        errorURL.searchParams.set("error", res.error);
        window.location.href = errorURL.toString();
        return;
      }

      window.location.href = getAuthCallbackPath(callbackUrl, isMobileApp);
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
          Verifying your secure sign-in {isMobileApp ? "code" : "link"}...
        </Text>
      </Flex>
    </Flex>
  );
};
