"use client";

import { useState } from "react";
import { Box, Button, Text } from "ui";
import { Logo } from "@/components/ui/logo";
import { getApiUrl } from "@/lib/api-url";

export const WorkspaceSSORecovery = ({
  workspaceSlug,
  workspaceName,
  email,
  required,
}: {
  workspaceSlug: string;
  workspaceName: string;
  email: string;
  required: boolean;
}) => {
  const [pending, setPending] = useState(false);
  return (
    <Box className="w-full max-w-md px-6">
      <Logo className="mb-8" />
      <Text as="h1" className="text-2xl font-medium">
        {required ? "Workspace SSO required" : "Connect workspace SSO"}
      </Text>
      <Text className="mt-3" color="muted">
        Sign in through {workspaceName}&apos;s identity provider to continue.
      </Text>
      <Box className="border-border bg-surface mt-6 rounded-xl border p-4">
        <Text fontWeight="medium">Your FortyOne account</Text>
        <Text className="mt-1 break-all" color="muted">
          {email}
        </Text>
      </Box>
      <Text className="mt-4" color="muted">
        Use the same email at your identity provider to connect this account.
      </Text>
      <Button
        align="center"
        className="mt-6"
        disabled={pending}
        fullWidth
        loading={pending}
        onClick={() => {
          setPending(true);
          window.location.assign(
            `${getApiUrl()}/auth/sso/${encodeURIComponent(workspaceSlug)}`,
          );
        }}
        size="lg"
      >
        {pending ? "Opening SSO…" : "Continue with SSO"}
      </Button>
      <Text className="mt-5" color="muted">
        Having trouble? Contact your workspace administrator.
      </Text>
    </Box>
  );
};
