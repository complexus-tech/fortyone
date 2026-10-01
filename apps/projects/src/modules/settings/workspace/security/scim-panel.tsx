"use client";

import { useId, useState } from "react";
import { Badge, Box, Button, Dialog, Flex, Input, Select, Text } from "ui";
import { getApiUrl } from "@/lib/api-url";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useSCIMMutations, useSCIMStatus } from "./scim-api";

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
    <Box className="border-border overflow-hidden rounded-xl border">
      <SectionHeader
        description="Sync workspace members with your identity provider."
        title="Member provisioning"
      />
      <Box className="space-y-5 px-6 py-5">
        <Box className="border-border bg-surface-elevated rounded-xl border p-4">
          <Text fontWeight="medium">SCIM base URL</Text>
          <Text className="mt-2 break-all select-all">{endpoint}</Text>
          <Text className="mt-2" color="muted">
            Supports users and deactivation. Team groups are managed in
            FortyOne.
          </Text>
        </Box>
        {status.isPending ? (
          <Text color="muted">Loading provisioning…</Text>
        ) : null}
        {status.isError ? (
          <Box>
            <Text color="danger" role="alert">
              Provisioning settings could not be loaded.
            </Text>
            <Button
              className="mt-3"
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
            <Flex align="center" className="gap-3" justify="between" wrap>
              <Text>{status.data.managedUsers} managed members</Text>
              <Button
                onClick={() => {
                  mutations.mint.reset();
                  setCreating(true);
                }}
              >
                Create token
              </Button>
            </Flex>
            {status.data.pendingSeatSync ? (
              <Box className="border-border rounded-xl border p-4">
                <Text fontWeight="medium">Seat sync pending</Text>
                <Text className="mt-1" color="muted">
                  {status.data.seatSyncError ||
                    "Membership changed. Sync the current seat count."}
                </Text>
                <Button
                  className="mt-3"
                  disabled={mutations.retry.isPending}
                  loading={mutations.retry.isPending}
                  onClick={() => {
                    mutations.retry.mutate();
                  }}
                  variant="outline"
                >
                  Retry seat sync
                </Button>
              </Box>
            ) : null}
            <Box className="divide-border divide-y">
              {status.data.credentials.length === 0 ? (
                <Text color="muted">
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
                      className="gap-4 py-4 first:pt-0"
                      justify="between"
                      key={credential.id}
                      wrap
                    >
                      <Box className="min-w-0">
                        <Text fontWeight="medium">{credential.name}</Text>
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
          <Text color="danger" role="alert">
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
        <Dialog.Content className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]">
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
                <Input
                  autoComplete="off"
                  label="Bearer token"
                  readOnly
                  type="text"
                  value={token}
                />
                <Button
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
                <Input
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
                  <label className="mb-1 block" htmlFor={lifetimeId}>
                    Expires after
                  </label>
                  <Select onValueChange={setLifetime} value={lifetime}>
                    <Select.Trigger
                      className="h-11 w-full px-4 text-base"
                      id={lifetimeId}
                    >
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
        <Dialog.Content className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]">
          <Dialog.Header className="shrink-0 px-6 py-5">
            <Dialog.Title className="text-lg">Revoke SCIM token?</Dialog.Title>
            <Dialog.Description className="px-0 text-base">
              The provider using this token will lose provisioning access.
            </Dialog.Description>
          </Dialog.Header>
          <Dialog.Footer className="shrink-0 justify-end gap-3 py-4">
            <Button
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
