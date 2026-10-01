"use client";

import { useState } from "react";
import { Button, Input, Text } from "ui";
import { getApiUrl } from "@/lib/api-url";

export const WorkspaceSSOSignIn = () => {
  const [workspace, setWorkspace] = useState("");
  const [pending, setPending] = useState(false);
  const valid = /^[a-zA-Z0-9][a-zA-Z0-9-]{0,119}$/.test(workspace.trim());
  return (
    <details className="border-border mt-4 border-t pt-4">
      <summary className="cursor-pointer font-medium">
        Sign in with workspace SSO
      </summary>
      <form
        className="mt-4 space-y-4"
        onSubmit={(event) => {
          event.preventDefault();
          if (valid) {
            setPending(true);
            window.location.assign(
              `${getApiUrl()}/auth/sso/${encodeURIComponent(workspace.trim())}`,
            );
          }
        }}
      >
        <Input
          autoComplete="organization"
          disabled={pending}
          label="Workspace URL name"
          maxLength={120}
          onChange={(event) => {
            setWorkspace(event.target.value);
          }}
          placeholder="acme"
          value={workspace}
        />
        <Text color="muted">
          Use the workspace name from your FortyOne URL.
        </Text>
        <Button
          align="center"
          color="tertiary"
          disabled={!valid || pending}
          fullWidth
          loading={pending}
          size="lg"
          type="submit"
        >
          {pending ? "Opening SSO…" : "Continue with SSO"}
        </Button>
      </form>
    </details>
  );
};
