"use client";

import { useState, type ReactNode } from "react";
import {
  CopyIcon,
  LockKeyholeIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshIcon,
} from "icons";
import { toast } from "sonner";
import { Box, Button, Dialog, Flex, Input, Menu, Text, TextArea } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useCopyToClipboard } from "@/hooks";
import { SectionHeader } from "@/modules/settings/components";
import { expiryFromNow, formatDeveloperDate } from "../constants";
import {
  useOAuthApplicationMutations,
  useOAuthApplications,
  useOAuthClientSecretMutations,
  useOAuthClientSecrets,
} from "../hooks";
import type {
  IssuedOAuthApplication,
  IssuedOAuthSecret,
  OAuthApplication,
} from "../types";
import { LoadError } from "./load-error";
import { SecretDialog } from "./secret-dialog";

const errorMessage = (error: unknown) =>
  error instanceof Error
    ? error.message
    : "The request could not be completed.";

const OAuthApplicationRow = ({
  application,
  onIssued,
}: {
  application: OAuthApplication;
  onIssued: (value: IssuedOAuthSecret) => void;
}) => {
  const [expanded, setExpanded] = useState(false);
  const [revokeId, setRevokeId] = useState<string | null>(null);
  const [, copy] = useCopyToClipboard();
  const secrets = useOAuthClientSecrets(application.id, expanded);
  const mutations = useOAuthClientSecretMutations(application.id);

  const rotate = async () => {
    try {
      const value = await mutations.rotate.mutateAsync(expiryFromNow(90));
      onIssued(value);
      toast.success("OAuth client secret rotated");
    } catch (error) {
      toast.error("Could not rotate client secret", {
        description: errorMessage(error),
      });
    }
  };

  const revoke = async () => {
    if (!revokeId) return;
    try {
      await mutations.revoke.mutateAsync(revokeId);
      setRevokeId(null);
      toast.success("OAuth client secret revoked");
    } catch (error) {
      toast.error("Could not revoke client secret", {
        description: errorMessage(error),
      });
    }
  };

  let secretContent: ReactNode = secrets.data?.map((secret) => (
    <Flex
      align="center"
      className="gap-3 px-4 py-3"
      justify="between"
      key={secret.id}
      wrap
    >
      <Box>
        <Text className="font-mono">Prefix {secret.prefix}</Text>
        <Text color="muted">
          Expires {formatDeveloperDate(secret.expiresAt)}
          {secret.lastUsedAt
            ? ` · Last used ${formatDeveloperDate(secret.lastUsedAt)}`
            : " · Never used"}
        </Text>
      </Box>
      {!secret.revokedAt ? (
        <Button
          color="tertiary"
          disabled={mutations.revoke.isPending}
          onClick={() => {
            setRevokeId(secret.id);
          }}
          size="sm"
          variant="naked"
        >
          Revoke
        </Button>
      ) : (
        <Text color="muted">Revoked</Text>
      )}
    </Flex>
  ));
  if (secrets.isLoading) {
    secretContent = (
      <Text className="px-4 py-4" color="muted">
        Loading client secrets…
      </Text>
    );
  } else if (secrets.isError) {
    secretContent = (
      <LoadError
        label="OAuth client secrets"
        onRetry={() => void secrets.refetch()}
      />
    );
  } else if ((secrets.data?.length ?? 0) === 0) {
    secretContent = (
      <Text className="px-4 py-4" color="muted">
        No client secrets remain for this application.
      </Text>
    );
  }

  return (
    <Box className="px-6 py-4">
      <Flex align="center" className="gap-4" justify="between" wrap>
        <Box className="min-w-0 flex-1">
          <Flex align="center" gap={2}>
            <Text className="font-medium">{application.name}</Text>
            <Text className="capitalize" color="muted">
              {application.status}
            </Text>
          </Flex>
          <Flex align="center" className="mt-1" gap={1}>
            <Text className="truncate font-mono" color="muted">
              {application.clientId}
            </Text>
            <Button
              aria-label="Copy client ID"
              color="tertiary"
              onClick={() => void copy(application.clientId)}
              size="sm"
              variant="naked"
            >
              <CopyIcon className="h-4" />
            </Button>
          </Flex>
          <Text color="muted">
            Expires {formatDeveloperDate(application.expiresAt)} ·{" "}
            {application.redirectUris.length} redirect{" "}
            {application.redirectUris.length === 1 ? "URI" : "URIs"}
          </Text>
        </Box>
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Actions for ${application.name}`}
              asIcon
              color="tertiary"
              disabled={mutations.rotate.isPending}
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
                {expanded ? "Hide secrets" : "Manage secrets"}
              </Menu.Item>
              <Menu.Item
                disabled={application.status !== "active"}
                onSelect={() => void rotate()}
              >
                <RefreshIcon className="h-[1.15rem]" />
                Rotate secret
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>

      {expanded ? (
        <Box className="border-border bg-surface-elevated mt-4 divide-y rounded-xl border">
          {secretContent}
        </Box>
      ) : null}

      <ConfirmDialog
        confirmPhrase="revoke"
        confirmText="Revoke secret"
        description="Clients using this secret will stop authenticating. Other active secrets are not affected."
        isLoading={mutations.revoke.isPending}
        isOpen={Boolean(revokeId)}
        onClose={() => {
          setRevokeId(null);
        }}
        onConfirm={() => void revoke()}
        title="Revoke OAuth client secret?"
      />
    </Box>
  );
};

export const OAuthApplications = () => {
  const applications = useOAuthApplications();
  const mutations = useOAuthApplicationMutations();
  const [createOpen, setCreateOpen] = useState(false);
  const [name, setName] = useState("");
  const [redirectURIs, setRedirectURIs] = useState("");
  const [applicationDays, setApplicationDays] = useState("365");
  const [secretDays, setSecretDays] = useState("90");
  const [issued, setIssued] = useState<{
    clientId: string;
    secret: string;
    description: string;
  } | null>(null);

  const create = async () => {
    const appDays = Number(applicationDays);
    const clientSecretDays = Number(secretDays);
    try {
      const value: IssuedOAuthApplication = await mutations.create.mutateAsync({
        name: name.trim(),
        redirectUris: redirectURIs
          .split("\n")
          .map((uri) => uri.trim())
          .filter(Boolean),
        expiresAt: expiryFromNow(appDays),
        secretExpiresAt: expiryFromNow(clientSecretDays),
      });
      setCreateOpen(false);
      setName("");
      setRedirectURIs("");
      setIssued({
        clientId: value.application.clientId,
        secret: value.secret,
        description:
          "Store the client secret now. Use the client ID in authorization requests; only the secret is confidential.",
      });
      toast.success("OAuth application created");
    } catch (error) {
      toast.error("Could not create OAuth application", {
        description: errorMessage(error),
      });
    }
  };

  const appDays = Number(applicationDays);
  const clientSecretDays = Number(secretDays);
  const isValid =
    name.trim().length > 0 &&
    Number.isInteger(appDays) &&
    appDays >= 1 &&
    appDays <= 365 &&
    Number.isInteger(clientSecretDays) &&
    clientSecretDays >= 1 &&
    clientSecretDays <= 365;

  let applicationContent: ReactNode = applications.data?.map((application) => (
    <OAuthApplicationRow
      application={application}
      key={application.id}
      onIssued={(value) => {
        setIssued({
          clientId: application.clientId,
          secret: value.secret,
          description: value.previousSecretOverlapExpiresAt
            ? `Update clients before ${formatDeveloperDate(
                value.previousSecretOverlapExpiresAt,
              )}. The previous secret remains valid until then.`
            : "Store this client secret now. It cannot be revealed again.",
        });
      }}
    />
  ));
  if (applications.isLoading) {
    applicationContent = (
      <Text className="px-6 py-5" color="muted">
        Loading OAuth applications…
      </Text>
    );
  } else if (applications.isError) {
    applicationContent = (
      <LoadError
        label="OAuth applications"
        onRetry={() => void applications.refetch()}
      />
    );
  } else if ((applications.data?.length ?? 0) === 0) {
    applicationContent = (
      <Text className="px-6 py-5" color="muted">
        No OAuth applications yet.
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
        description="Register OAuth 2.0 clients for delegated user access or approved application-to-application workflows."
        title="OAuth applications"
      />
      <Box className="divide-border divide-y-[0.5px]">{applicationContent}</Box>

      <Dialog onOpenChange={setCreateOpen} open={createOpen}>
        <Dialog.Content className="max-w-3xl" size="lg">
          <Dialog.Header className="px-6 pt-6 pb-2">
            <Dialog.Title className="text-xl">
              Create OAuth application
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="space-y-5 px-6 pt-2 pb-4">
            <Input
              autoFocus
              label="Application name"
              maxLength={120}
              onChange={(event) => {
                setName(event.target.value);
              }}
              placeholder="Reporting integration"
              value={name}
            />
            <Box>
              <TextArea
                label="Redirect URIs"
                onChange={(event) => {
                  setRedirectURIs(event.target.value);
                }}
                placeholder={
                  "https://example.com/oauth/callback\nhttps://staging.example.com/oauth/callback"
                }
                rows={4}
                value={redirectURIs}
              />
              <Text className="mt-1" color="muted">
                One exact HTTPS URI per line. Leave empty only for approved
                client-credentials workflows.
              </Text>
            </Box>
            <Flex className="flex-col sm:flex-row" gap={4}>
              <Box className="flex-1">
                <Input
                  label="Application lifetime (days)"
                  max={365}
                  min={1}
                  onChange={(event) => {
                    setApplicationDays(event.target.value);
                  }}
                  type="number"
                  value={applicationDays}
                />
              </Box>
              <Box className="flex-1">
                <Input
                  label="Secret lifetime (days)"
                  max={365}
                  min={1}
                  onChange={(event) => {
                    setSecretDays(event.target.value);
                  }}
                  type="number"
                  value={secretDays}
                />
              </Box>
            </Flex>
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
              disabled={!isValid}
              loading={mutations.create.isPending}
              onClick={() => void create()}
            >
              Create application
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>

      <SecretDialog
        description={issued?.description ?? ""}
        label="Client secret"
        onClose={() => {
          setIssued(null);
        }}
        open={Boolean(issued)}
        secret={issued?.secret ?? ""}
        title="OAuth client credentials"
        values={issued ? [{ label: "Client ID", value: issued.clientId }] : []}
      />
    </Box>
  );
};
