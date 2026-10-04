"use client";

import { useId, useState } from "react";
import { Badge, Box, Button, Dialog, Flex, Select, Text } from "ui";
import { getApiUrl } from "@/lib/api-url";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useSCIMMutations, useSCIMStatus } from "./scim-api";
import { SecuritySettingRow } from "./setting-row";
import { SecurityInput } from "./security-input";

export const SecurityProvisioning = () => {
  const status = useSCIMStatus();
  const mutations = useSCIMMutations();
  const { workspaceSlug } = useWorkspacePath();
  const lifetimeId = useId();
  const [creating, setCreating] = useState(false);
  const [name, setName] = useState("");
  const [lifetime, setLifetime] = useState("90");
  const [token, setToken] = useState("");
  const [revokeId, setRevokeId] = useState("");
  const [copyError, setCopyError] = useState("");
  const [copied, setCopied] = useState(false);
  const endpoint = `${getApiUrl()}/scim/v2/${encodeURIComponent(workspaceSlug)}`;
  const error =
    mutations.mint.error ?? mutations.revoke.error ?? mutations.retry.error;
  const closeCreate = () => {
    setCreating(false);
    setToken("");
    setName("");
    setCopyError("");
    setCopied(false);
    mutations.mint.reset();
  };
  return (
    <Box className="border-border bg-surface overflow-hidden rounded-2xl border">
      <SectionHeader
        action={
          status.isSuccess ? (
            <Button
              color="tertiary"
              onClick={() => {
                mutations.mint.reset();
                setCreating(true);
              }}
              variant="outline"
            >
              Create token
            </Button>
          ) : null
        }
        description="Sync workspace members with your identity provider."
        title="Member provisioning"
      />
      <Box className="divide-border divide-y-[0.5px]">
        <SecuritySettingRow
          description="Supports users and deactivation. Team groups are managed in FortyOne."
          title="SCIM base URL"
        >
          <Text className="max-w-full break-all select-all md:max-w-[60%] md:shrink-0">
            {endpoint}
          </Text>
        </SecuritySettingRow>
        {status.isPending ? (
          <Text aria-live="polite" className="px-6 py-4" color="muted">
            Loading provisioning…
          </Text>
        ) : null}
        {status.isError ? (
          <Box className="px-6 py-4">
            <Text color="danger" role="alert">
              Provisioning settings could not be loaded.
            </Text>
            <Button
              className="mt-3"
              color="tertiary"
              onClick={() => {
                void status.refetch();
              }}
              variant="outline"
            >
              Try again
            </Button>
          </Box>
        ) : null}
        {status.isSuccess ? (
          <>
            <SecuritySettingRow
              description="Members managed by your identity provider."
              layout="inline"
              title="Provisioned members"
            >
              <Text className="shrink-0">
                {status.data.managedUsers} managed members
              </Text>
            </SecuritySettingRow>
            {status.data.pendingSeatSync ? (
              <SecuritySettingRow
                description={
                  status.data.seatSyncError ||
                  "Membership changed. Sync the current seat count."
                }
                title="Seat sync pending"
              >
                <Button
                  className="shrink-0"
                  color="tertiary"
                  disabled={mutations.retry.isPending}
                  loading={mutations.retry.isPending}
                  loadingText="Syncing seats…"
                  onClick={() => {
                    mutations.retry.mutate();
                  }}
                  variant="outline"
                >
                  Retry seat sync
                </Button>
              </SecuritySettingRow>
            ) : null}
            <Box className="divide-border divide-y-[0.5px]">
              {status.data.credentials.length === 0 ? (
                <Text className="px-6 py-4" color="muted">
                  Create a token to connect your provider.
                </Text>
              ) : (
                status.data.credentials.map((credential) => {
                  const unavailable =
                    Boolean(credential.revokedAt) ||
                    new Date(credential.expiresAt).getTime() <= Date.now();
                  return (
                    <Flex
                      align="center"
                      className="gap-4 px-6 py-4"
                      justify="between"
                      key={credential.id}
                      wrap
                    >
                      <Box className="min-w-0 flex-1">
                        <Text className="break-words" fontWeight="medium">
                          {credential.name}
                        </Text>
                        <Text className="mt-1 break-all" color="muted">
                          {credential.prefix}… · Expires{" "}
                          {new Date(credential.expiresAt).toLocaleDateString()}
                        </Text>
                      </Box>
                      {unavailable ? (
                        <Badge color="tertiary" variant="outline">
                          {credential.revokedAt ? "Revoked" : "Expired"}
                        </Badge>
                      ) : (
                        <Button
                          color="danger"
                          onClick={() => {
                            mutations.revoke.reset();
                            setRevokeId(credential.id);
                          }}
                          variant="outline"
                        >
                          Revoke
                        </Button>
                      )}
                    </Flex>
                  );
                })
              )}
            </Box>
          </>
        ) : null}
        {error ? (
          <Text className="px-6 py-4" color="danger" role="alert">
            {error.message}
          </Text>
        ) : null}
      </Box>
      <Dialog
        onOpenChange={(open) => {
          if (!open && !mutations.mint.isPending) closeCreate();
        }}
        open={creating}
      >
        <Dialog.Content
          aria-busy={mutations.mint.isPending}
          className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
          hideClose={mutations.mint.isPending}
          onEscapeKeyDown={(event) => {
            if (mutations.mint.isPending) event.preventDefault();
          }}
          onInteractOutside={(event) => {
            if (mutations.mint.isPending) event.preventDefault();
          }}
        >
          <Dialog.Header className="shrink-0 px-6 py-5">
            <Dialog.Title className="text-lg">
              {token ? "Copy your SCIM token" : "Create SCIM token"}
            </Dialog.Title>
            <Dialog.Description className="px-0 text-base">
              {token
                ? "This token is shown once. Save it in your provider."
                : "Use a separate token for each identity provider."}
            </Dialog.Description>
          </Dialog.Header>
          <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-4 py-5">
            {token ? (
              <>
                <SecurityInput
                  autoComplete="off"
                  label="Bearer token"
                  readOnly
                  type="text"
                  value={token}
                />
                <Button
                  color="tertiary"
                  onClick={() => {
                    void navigator.clipboard
                      .writeText(token)
                      .then(() => {
                        setCopied(true);
                      })
                      .catch(() => {
                        setCopyError("Copy failed. Select and copy the token.");
                      });
                  }}
                  variant="outline"
                >
                  {copied ? "Copied" : "Copy token"}
                </Button>
                {copyError ? (
                  <Text color="danger" role="alert">
                    {copyError}
                  </Text>
                ) : null}
              </>
            ) : (
              <>
                <SecurityInput
                  disabled={mutations.mint.isPending}
                  label="Token name"
                  maxLength={100}
                  onChange={(event) => {
                    setName(event.target.value);
                  }}
                  placeholder="Identity provider"
                  value={name}
                />
                <Box>
                  <label className="mb-[0.35rem] block" htmlFor={lifetimeId}>
                    Expires after
                  </label>
                  <Select
                    disabled={mutations.mint.isPending}
                    onValueChange={setLifetime}
                    value={lifetime}
                  >
                    <Select.Trigger className="text-base" id={lifetimeId}>
                      <Select.Input />
                    </Select.Trigger>
                    <Select.Content>
                      {[
                        ["30", "30 days"],
                        ["90", "90 days"],
                        ["365", "1 year"],
                      ].map(([value, label]) => (
                        <Select.Option
                          className="py-2 text-base"
                          key={value}
                          value={value}
                        >
                          {label}
                        </Select.Option>
                      ))}
                    </Select.Content>
                  </Select>
                </Box>
                {mutations.mint.error ? (
                  <Text color="danger" role="alert">
                    {mutations.mint.error.message}
                  </Text>
                ) : null}
              </>
            )}
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 justify-end gap-3 py-4">
            <Button
              color="tertiary"
              disabled={mutations.mint.isPending}
              onClick={closeCreate}
              variant="outline"
            >
              {token ? "Done" : "Cancel"}
            </Button>
            {!token ? (
              <Button
                disabled={!name.trim() || mutations.mint.isPending}
                loading={mutations.mint.isPending}
                loadingText="Creating token…"
                onClick={() => {
                  mutations.mint.mutate(
                    { name: name.trim(), lifetimeDays: Number(lifetime) },
                    {
                      onSuccess: (result) => {
                        setToken(result.token);
                        mutations.mint.reset();
                      },
                    },
                  );
                }}
              >
                Create token
              </Button>
            ) : null}
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
      <Dialog
        onOpenChange={(open) => {
          if (!open && !mutations.revoke.isPending) setRevokeId("");
        }}
        open={Boolean(revokeId)}
      >
        <Dialog.Content
          aria-busy={mutations.revoke.isPending}
          className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
          hideClose={mutations.revoke.isPending}
          onEscapeKeyDown={(event) => {
            if (mutations.revoke.isPending) event.preventDefault();
          }}
          onInteractOutside={(event) => {
            if (mutations.revoke.isPending) event.preventDefault();
          }}
        >
          <Dialog.Header className="shrink-0 px-6 py-5">
            <Dialog.Title className="text-lg">Revoke SCIM token?</Dialog.Title>
            <Dialog.Description className="px-0 text-base">
              The provider using this token will lose provisioning access.
            </Dialog.Description>
          </Dialog.Header>
          {mutations.revoke.error ? (
            <Dialog.Body>
              <Text color="danger" role="alert">
                {mutations.revoke.error.message}
              </Text>
            </Dialog.Body>
          ) : null}
          <Dialog.Footer className="shrink-0 justify-end gap-3 py-4">
            <Button
              color="tertiary"
              disabled={mutations.revoke.isPending}
              onClick={() => {
                setRevokeId("");
              }}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              color="danger"
              disabled={mutations.revoke.isPending}
              loading={mutations.revoke.isPending}
              loadingText="Revoking token…"
              onClick={() => {
                mutations.revoke.mutate(revokeId, {
                  onSuccess: () => {
                    setRevokeId("");
                  },
                });
              }}
            >
              Revoke token
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};
