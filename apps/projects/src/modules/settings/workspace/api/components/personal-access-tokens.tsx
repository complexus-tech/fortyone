"use client";

import { useState, type ReactNode } from "react";
import { PlusIcon } from "icons";
import { toast } from "sonner";
import { Box, Button, Text } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { SectionHeader } from "@/modules/settings/components";
import { expiryFromNow } from "../constants";
import { usePersonalTokenMutations, usePersonalTokens } from "../hooks";
import type { CreateCredentialInput, IssuedCredential } from "../types";
import { CredentialDialog } from "./credential-dialog";
import { CredentialRow } from "./credential-row";
import { LoadError } from "./load-error";
import { SecretDialog } from "./secret-dialog";

const errorMessage = (error: unknown) =>
  error instanceof Error
    ? error.message
    : "The request could not be completed.";

export const PersonalAccessTokens = () => {
  const tokens = usePersonalTokens();
  const mutations = usePersonalTokenMutations();
  const [createOpen, setCreateOpen] = useState(false);
  const [issued, setIssued] = useState<IssuedCredential | null>(null);
  const [rotateId, setRotateId] = useState<string | null>(null);
  const [revokeId, setRevokeId] = useState<string | null>(null);
  const isPending =
    mutations.create.isPending ||
    mutations.rotate.isPending ||
    mutations.revoke.isPending;

  const create = async (input: CreateCredentialInput) => {
    try {
      const value = await mutations.create.mutateAsync(input);
      setCreateOpen(false);
      setIssued(value);
      toast.success("Personal access token created");
    } catch (error) {
      toast.error("Could not create token", {
        description: errorMessage(error),
      });
      throw error;
    }
  };

  const rotate = async () => {
    if (!rotateId) return;
    try {
      const value = await mutations.rotate.mutateAsync({
        credentialId: rotateId,
        expiresAt: expiryFromNow(90),
      });
      setRotateId(null);
      setIssued(value);
      toast.success("Personal access token rotated");
    } catch (error) {
      toast.error("Could not rotate token", {
        description: errorMessage(error),
      });
    }
  };

  const revoke = async () => {
    if (!revokeId) return;
    try {
      await mutations.revoke.mutateAsync(revokeId);
      setRevokeId(null);
      toast.success("Personal access token revoked");
    } catch (error) {
      toast.error("Could not revoke token", {
        description: errorMessage(error),
      });
    }
  };

  let tokenContent: ReactNode = tokens.data?.map((token) => (
    <CredentialRow
      credential={token}
      isPending={isPending}
      key={token.id}
      onRevoke={() => {
        setRevokeId(token.id);
      }}
      onRotate={() => {
        setRotateId(token.id);
      }}
    />
  ));
  if (tokens.isLoading) {
    tokenContent = (
      <Text className="px-6 py-5" color="muted">
        Loading tokens…
      </Text>
    );
  } else if (tokens.isError) {
    tokenContent = (
      <LoadError
        label="personal access tokens"
        onRetry={() => void tokens.refetch()}
      />
    );
  } else if ((tokens.data?.length ?? 0) === 0) {
    tokenContent = (
      <Text className="px-6 py-5" color="muted">
        No personal access tokens yet.
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
        description="Create scoped credentials for your own scripts and command-line tools. Tokens are shown once and expire automatically."
        title="Personal access tokens"
      />
      <Box className="divide-border divide-y-[0.5px]">{tokenContent}</Box>

      <CredentialDialog
        description="Give this token only the permissions and team access your tool needs."
        isPending={mutations.create.isPending}
        onOpenChange={setCreateOpen}
        onSubmit={create}
        open={createOpen}
        title="Create personal access token"
      />
      <SecretDialog
        description="Store this token in a password manager or secret store. FortyOne keeps only a one-way digest and cannot reveal it again."
        label="Access token"
        onClose={() => {
          setIssued(null);
        }}
        open={Boolean(issued)}
        secret={issued?.token ?? ""}
        title="Your token is ready"
      />
      <ConfirmDialog
        confirmText="Rotate token"
        description="The current token will stop working immediately. Be ready to replace it everywhere the token is used."
        isLoading={mutations.rotate.isPending}
        isOpen={Boolean(rotateId)}
        onClose={() => {
          setRotateId(null);
        }}
        onConfirm={() => void rotate()}
        title="Rotate personal access token?"
      />
      <ConfirmDialog
        confirmPhrase="revoke"
        confirmText="Revoke token"
        description="This token will stop working immediately. This action cannot be undone."
        isLoading={mutations.revoke.isPending}
        isOpen={Boolean(revokeId)}
        onClose={() => {
          setRevokeId(null);
        }}
        onConfirm={() => void revoke()}
        title="Revoke personal access token?"
      />
    </Box>
  );
};
