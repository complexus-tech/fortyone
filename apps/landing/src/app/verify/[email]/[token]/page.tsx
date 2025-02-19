"use client";

import { Text, Flex } from "ui";
import { redirect, useParams } from "next/navigation";
import { useEffect } from "react";
import { Logo, Blur } from "@/components/ui";
import { getRedirectUrl } from "@/utils";
import { logIn, getSession } from "./actions";

export default function EmailVerificationCallback() {
  const params = useParams<{ email: string; token: string }>();
  const validatedEmail = decodeURIComponent(params?.email || "");
  const validatedToken = decodeURIComponent(params?.token || "");

  useEffect(() => {
    const validate = async () => {
      const res = await logIn(validatedEmail, validatedToken);
      if (res?.error) {
        redirect(`/login?error=${res.error}`);
      } else {
        const session = await getSession();
        if (session) {
          redirect(getRedirectUrl(session));
        }
      }
    };

    void validate();
  }, [validatedEmail, validatedToken]);

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
          Verifying signin link...
        </Text>
      </Flex>
    </Flex>
  );
}
