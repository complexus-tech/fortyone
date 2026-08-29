"use client";

import { useState, type ReactNode } from "react";
import {
  DeleteIcon,
  LockKeyholeIcon,
  MoreHorizontalIcon,
  PlusIcon,
} from "icons";
import { toast } from "sonner";
import { Box, Button, Dialog, Flex, Input, Menu, Text } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { SectionHeader } from "@/modules/settings/components";
import { expiryFromNow } from "../constants";
import {
  useServiceAccountKeyMutations,
  useServiceAccountKeys,
  useServiceAccountMutations,
  useServiceAccounts,
} from "../hooks";
import type {
  CreateCredentialInput,
  IssuedCredential,
  ServiceAccount,
} from "../types";
import { CredentialDialog } from "./credential-dialog";
import { CredentialRow } from "./credential-row";
import { LoadError } from "./load-error";
import { SecretDialog } from "./secret-dialog";

const errorMessage = (error: unknown) =>
  error instanceof Error
    ? error.message
    : "The request could not be completed.";

const ServiceAccountRow = ({
  account,
  onDisable,
}: {
  account: ServiceAccount;
  onDisable: () => void;
}) => {
  const [expanded, setExpanded] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [issued, setIssued] = useState<IssuedCredential | null>(null);
  const [revokeId, setRevokeId] = useState<string | null>(null);
  const keys = useServiceAccountKeys(account.id, expanded);
  const mutations = useServiceAccountKeyMutations(account.id);
  const isPending =
    mutations.create.isPending ||
    mutations.rotate.isPending ||
    mutations.revoke.isPending;

  const create = async (input: CreateCredentialInput) => {
    try {
      const value = await mutations.create.mutateAsync(input);
      setCreateOpen(false);
      setIssued(value);
      toast.success("Service-account key created");
    } catch (error) {
      toast.error("Could not create key", { description: errorMessage(error) });
      throw error;
    }
  };

  const rotate = async (credentialId: string) => {
    try {
      const value = await mutations.rotate.mutateAsync({
        credentialId,
        expiresAt: expiryFromNow(90),
      });
      setIssued(value);
      toast.success("Service-account key rotated");
    } catch (error) {
      toast.error("Could not rotate key", { description: errorMessage(error) });
    }
  };

  const revoke = async () => {
    if (!revokeId) return;
    try {
      await mutations.revoke.mutateAsync(revokeId);
      setRevokeId(null);
      toast.success("Service-account key revoked");
    } catch (error) {
      toast.error("Could not revoke key", { description: errorMessage(error) });
    }
  };

  let keyContent: ReactNode = keys.data?.map((key) => (
    <CredentialRow
      credential={key}
      isPending={isPending}
      key={key.id}
      onRevoke={() => {
        setRevokeId(key.id);
      }}
      onRotate={() => void rotate(key.id)}
    />
  ));
  if (keys.isLoading) {
    keyContent = (
      <Text className="px-4 py-4" color="muted">
        Loading keys…
      </Text>
    );
  } else if (keys.isError) {
    keyContent = (
      <LoadError
        label="service-account keys"
        onRetry={() => void keys.refetch()}
      />
    );
  } else if ((keys.data?.length ?? 0) === 0) {
    keyContent = (
      <Text className="px-4 py-4" color="muted">
        No keys created for this service account.
      </Text>
    );
  }

  return (
    <Box className="px-6 py-4">
      <Flex align="center" className="gap-4" justify="between" wrap>
        <Box className="min-w-0 flex-1">
          <Flex align="center" gap={2}>
            <Text className="font-medium">{account.name}</Text>
            <Text className="capitalize" color="muted">
              {account.status}
            </Text>
          </Flex>
          <Text className="mt-1 capitalize" color="muted">
            Workspace role: {account.workspaceRole}
          </Text>
        </Box>
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Actions for ${account.name}`}
              asIcon
              color="tertiary"
              leftIcon={<MoreHorizontalIcon />}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">More options</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  setExpanded((current) => !current);
                }}
              >
                <LockKeyholeIcon className="h-[1.15rem]" />
                {expanded ? "Hide keys" : "Manage keys"}
              </Menu.Item>
              <Menu.Item
                disabled={account.status !== "active"}
                onSelect={() => {
                  setCreateOpen(true);
                }}
              >
                <PlusIcon className="h-[1.15rem]" />
                Create key
              </Menu.Item>
              {account.status === "active" ? (
                <Menu.Item onSelect={onDisable}>
                  <DeleteIcon className="h-[1.15rem]" />
                  Disable
                </Menu.Item>
              ) : null}
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>

      {expanded ? (
        <Box className="border-border bg-surface-elevated mt-4 rounded-xl border">
          <Box className="divide-border divide-y-[0.5px]">{keyContent}</Box>
        </Box>
      ) : null}

      <CredentialDialog
        description={`Create a least-privilege key for ${account.name}. Keys authenticate as the service account, not as a human user.`}
        isPending={mutations.create.isPending}
        onOpenChange={setCreateOpen}
        onSubmit={create}
        open={createOpen}
        title="Create service-account key"
      />
      <SecretDialog
        description="Store this key now. FortyOne cannot reveal it again. The previous key remains valid for one hour after rotation."
        label="Service-account key"
        onClose={() => {
          setIssued(null);
        }}
        open={Boolean(issued)}
        secret={issued?.token ?? ""}
        title="Your service-account key is ready"
      />
      <ConfirmDialog
        confirmPhrase="revoke"
        confirmText="Revoke key"
        description="This key will stop working immediately. Other active keys on the service account are not affected."
        isLoading={mutations.revoke.isPending}
        isOpen={Boolean(revokeId)}
        onClose={() => {
          setRevokeId(null);
        }}
        onConfirm={() => void revoke()}
        title="Revoke service-account key?"
      />
    </Box>
  );
};

export const ServiceAccounts = () => {
  const accounts = useServiceAccounts();
  const mutations = useServiceAccountMutations();
  const [createOpen, setCreateOpen] = useState(false);
  const [name, setName] = useState("");
  const [workspaceRole, setWorkspaceRole] = useState<"guest" | "member">(
    "member",
  );
  const [disableId, setDisableId] = useState<string | null>(null);

  const create = async () => {
    try {
      await mutations.create.mutateAsync({
        name: name.trim(),
        workspaceRole,
      });
      setCreateOpen(false);
      setName("");
      setWorkspaceRole("member");
      toast.success("Service account created");
    } catch (error) {
      toast.error("Could not create service account", {
        description: errorMessage(error),
      });
    }
  };

  const disable = async () => {
    if (!disableId) return;
    try {
      await mutations.disable.mutateAsync(disableId);
      setDisableId(null);
      toast.success("Service account disabled");
    } catch (error) {
      toast.error("Could not disable service account", {
        description: errorMessage(error),
      });
    }
  };

  let accountContent: ReactNode = accounts.data?.map((account) => (
    <ServiceAccountRow
      account={account}
      key={account.id}
      onDisable={() => {
        setDisableId(account.id);
      }}
    />
  ));
  if (accounts.isLoading) {
    accountContent = (
      <Text className="px-6 py-5" color="muted">
        Loading service accounts…
      </Text>
    );
  } else if (accounts.isError) {
    accountContent = (
      <LoadError
        label="service accounts"
        onRetry={() => void accounts.refetch()}
      />
    );
  } else if ((accounts.data?.length ?? 0) === 0) {
    accountContent = (
      <Text className="px-6 py-5" color="muted">
        No service accounts yet.
      </Text>
    );
  }

  return (
    <Box className="border-border bg-surface rounded-2xl border">
      <SectionHeader
        action={
          <Button
            color="tertiary"
            leftIcon={<PlusIcon className="h-4" />}
            onClick={() => {
              setCreateOpen(true);
            }}
            size="sm"
          >
            Create
          </Button>
        }
        description="Use non-human identities for production automations. Each account has an explicit role and independently rotatable keys."
        title="Service accounts"
      />
      <Box className="divide-border divide-y-[0.5px]">{accountContent}</Box>

      <Dialog onOpenChange={setCreateOpen} open={createOpen}>
        <Dialog.Content className="max-w-2xl" size="md">
          <Dialog.Header className="px-6 pt-6 pb-2">
            <Dialog.Title className="text-xl">
              Create service account
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="space-y-5 px-6 pt-2 pb-4">
            <Text color="muted">
              Create the identity first, then issue one or more scoped keys.
            </Text>
            <Input
              autoFocus
              label="Name"
              maxLength={120}
              onChange={(event) => {
                setName(event.target.value);
              }}
              placeholder="Deployment automation"
              value={name}
            />
            <Box>
              <Text className="mb-2" fontWeight="medium">
                Workspace role
              </Text>
              <Flex gap={2}>
                {(["guest", "member"] as const).map((role) => (
                  <Button
                    aria-pressed={workspaceRole === role}
                    color={workspaceRole === role ? "primary" : "tertiary"}
                    key={role}
                    onClick={() => {
                      setWorkspaceRole(role);
                    }}
                    variant={workspaceRole === role ? "solid" : "outline"}
                  >
                    <span className="capitalize">{role}</span>
                  </Button>
                ))}
              </Flex>
            </Box>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              disabled={mutations.create.isPending}
              onClick={() => {
                setCreateOpen(false);
              }}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              disabled={!name.trim()}
              loading={mutations.create.isPending}
              onClick={() => void create()}
            >
              Create
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>

      <ConfirmDialog
        confirmPhrase="disable"
        confirmText="Disable account"
        description="Every key belonging to this service account will be revoked immediately."
        isLoading={mutations.disable.isPending}
        isOpen={Boolean(disableId)}
        onClose={() => {
          setDisableId(null);
        }}
        onConfirm={() => void disable()}
        title="Disable service account?"
      />
    </Box>
  );
};
