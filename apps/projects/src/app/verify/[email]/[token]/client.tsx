"use client";

import { Text, Flex } from "ui";
import { useParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { Logo } from "@/components/ui";
import { logIn } from "./actions";

export const EmailVerificationCallback = ({
  isMobileApp,
}: {
  isMobileApp: boolean;
}) => {
  const params = useParams<{ email: string; token: string }>();
  const validatedEmail = decodeURIComponent(params?.email || "");
  const validatedToken = decodeURIComponent(params?.token || "");
  const hasValidated = useRef(false);

  useEffect(() => {
    const validate = async () => {
      if (hasValidated.current) {
        return;
      }
      hasValidated.current = true;
      const res = await logIn(validatedEmail, validatedToken);

      if (res?.error) {
        window.location.href = `/?error=${encodeURIComponent(res.error)}`;
        return;
      }

      window.location.href = isMobileApp
        ? "/auth-callback?mobileApp=true"
        : "/auth-callback";
    };

    void validate();
  }, [validatedEmail, validatedToken, isMobileApp]);

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
