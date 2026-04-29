"use client";

import { useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { Flex, Text } from "ui";
import { Logo } from "@/components/ui";

const fallbackReturnTo = "/";

const isSafeReturnTo = (returnTo: string) => {
  if (returnTo.startsWith("/") && !returnTo.startsWith("//")) {
    return true;
  }
  try {
    const url = new URL(returnTo);
    if (url.protocol !== "http:" && url.protocol !== "https:") {
      return false;
    }
    if (url.hostname === window.location.hostname) {
      return true;
    }
    return url.hostname.endsWith(".fortyone.app");
  } catch {
    return false;
  }
};

const decodeReturnPath = (state: string | null) => {
  if (!state) {
    return fallbackReturnTo;
  }

  try {
    const [payload] = state.split(".");
    if (!payload) {
      return fallbackReturnTo;
    }
    const base64 = payload.replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=");
    const parsed = JSON.parse(atob(padded)) as { return_to?: string };
    const returnTo = parsed.return_to ?? fallbackReturnTo;
    if (!isSafeReturnTo(returnTo)) {
      return fallbackReturnTo;
    }
    return returnTo;
  } catch {
    return fallbackReturnTo;
  }
};

export default function GitHubCallbackPage() {
  const searchParams = useSearchParams();

  useEffect(() => {
    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const returnTo = decodeReturnPath(state);

    if (code && state) {
      const separator = returnTo.includes("?") ? "&" : "?";
      window.location.href = `${returnTo}${separator}code=${encodeURIComponent(code)}&github_state=${encodeURIComponent(state)}&github_link=1`;
    } else {
      window.location.href = returnTo;
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
