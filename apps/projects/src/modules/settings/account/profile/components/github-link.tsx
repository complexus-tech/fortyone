"use client";

import { useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { Box, Button, Flex, Text } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import {
  useLinkGitHubUser,
  useUnlinkGitHubUser,
} from "@/lib/hooks/github";
import { useProfile } from "@/lib/hooks/profile";

const GITHUB_CLIENT_ID = process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID ?? "";
const GITHUB_CALLBACK_URL = process.env.NEXT_PUBLIC_GITHUB_CALLBACK_URL ?? "";

// Module-level flag survives React Strict Mode double-mount.
// Prevents the OAuth code from being exchanged twice (GitHub codes are single-use).
let codeConsumed = false;

export const GitHubAccountLink = () => {
  const searchParams = useSearchParams();
  const linkGitHub = useLinkGitHubUser();
  const unlinkGitHub = useUnlinkGitHubUser();
  const { data: profile } = useProfile();

  const isLinked = !!profile?.githubUsername;

  useEffect(() => {
    const code = searchParams.get("code");
    const isGitHubLink = searchParams.get("github_link");
    if (code && isGitHubLink && !codeConsumed) {
      codeConsumed = true;
      linkGitHub.mutate(code);
      // Clean the code from the URL
      const url = new URL(window.location.href);
      url.searchParams.delete("code");
      url.searchParams.delete("github_link");
      window.history.replaceState({}, "", url.toString());
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  const handleConnect = () => {
    // Use a fixed callback URL so it works across workspace subdomains.
    // The return URL is base64-encoded in the OAuth `state` parameter — GitHub
    // passes it back untouched, so it survives cross-origin redirects.
    const returnUrl = btoa(window.location.href);
    const callbackUrl = GITHUB_CALLBACK_URL || `${window.location.origin}/github/callback`;
    window.location.href = `https://github.com/login/oauth/authorize?client_id=${GITHUB_CLIENT_ID}&redirect_uri=${encodeURIComponent(callbackUrl)}&state=${encodeURIComponent(returnUrl)}`;
  };

  const handleDisconnect = () => {
    unlinkGitHub.mutate();
  };

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description="Link your GitHub account to attribute pull requests, reviews, and commits to your FortyOne profile."
        title="GitHub Account"
      />
      <Box className="px-6 py-4">
        {isLinked ? (
          <Flex align="center" justify="between">
            <Flex align="center" gap={2}>
              <Text color="muted">
                Connected as{" "}
                <Text as="span" className="font-medium">
                  @{profile.githubUsername}
                </Text>
              </Text>
            </Flex>
            <Button
              color="danger"
              variant="naked"
              onClick={handleDisconnect}
              disabled={unlinkGitHub.isPending}
            >
              Disconnect
            </Button>
          </Flex>
        ) : (
          <Flex align="center" justify="between">
            <Text color="muted">
              Connect your GitHub account so your activity is attributed to you.
            </Text>
            <Button
              color="invert"
              onClick={handleConnect}
              disabled={!GITHUB_CLIENT_ID || linkGitHub.isPending}
            >
              {linkGitHub.isPending ? "Connecting..." : "Connect GitHub"}
            </Button>
          </Flex>
        )}
      </Box>
    </Box>
  );
};
