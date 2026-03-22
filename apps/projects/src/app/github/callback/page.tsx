"use client";

import { useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { Flex, Text } from "ui";
import { Logo } from "@/components/ui";

export default function GitHubCallbackPage() {
  const searchParams = useSearchParams();

  useEffect(() => {
    const code = searchParams.get("code");
    const state = searchParams.get("state");

    // Decode the return URL from the state parameter
    let returnTo: string | null = null;
    if (state) {
      try {
        returnTo = atob(state);
      } catch {
        // invalid base64, ignore
      }
    }

    if (code && returnTo) {
      const separator = returnTo.includes("?") ? "&" : "?";
      window.location.href = `${returnTo}${separator}code=${code}&github_link=1`;
    } else if (returnTo) {
      window.location.href = returnTo;
    } else {
      window.location.href = "/";
    }
  }, [searchParams]);

  return (
    <Flex align="center" className="h-dvh dark:bg-black" justify="center">
      <Flex align="center" direction="column" justify="center">
        <Logo asIcon className="animate-pulse" />
        <Text color="muted" fontWeight="medium" className="mt-4">
          Connecting your GitHub account...
        </Text>
      </Flex>
    </Flex>
  );
}
