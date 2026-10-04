"use client";

import { useState } from "react";
import { Badge, Box, Button, Dialog, Flex, Switch, Text } from "ui";
import { toast } from "sonner";
import { SectionHeader } from "@/modules/settings/components/section-header";
import type { SSOSettings } from "./sso-api";
import { useSSOMutations, useSSOSettings } from "./sso-api";
import { SecuritySettingRow } from "./setting-row";
import { SecurityInput } from "./security-input";

const ConnectionForm = ({ settings }: { settings: SSOSettings }) => {
  const connection = settings.connection;
  const [issuer, setIssuer] = useState("");
  const [clientId, setClientId] = useState("");
  const [secret, setSecret] = useState("");
  const [enabled, setEnabled] = useState(connection?.enabled ?? true);
  const [required, setRequired] = useState(connection?.requireSSO ?? false);
  const [rotating, setRotating] = useState(false);
  const [removing, setRemoving] = useState(false);
  const [reason, setReason] = useState("");
  const mutations = useSSOMutations();
  const busy =
    mutations.create.isPending ||
    mutations.update.isPending ||
    mutations.archive.isPending;
  const error =
    mutations.create.error ?? mutations.update.error ?? mutations.archive.error;
  const createValid =
    issuer.trim().startsWith("https://") &&
    clientId.trim().length > 0 &&
    secret.length > 0;
  const changed =
    connection &&
    (enabled !== connection.enabled ||
      required !== connection.requireSSO ||
      rotating);
  return (
    <Box className="divide-border divide-y-[0.5px]">
      <SecuritySettingRow
        description="Add this callback URL to your identity provider."
        title="Identity provider callback URL"
      >
        <Text className="max-w-full break-all select-all md:max-w-[60%] md:shrink-0">
          {settings.callbackUrl}
        </Text>
      </SecuritySettingRow>
      {!connection ? (
        <form
          className="divide-border divide-y-[0.5px]"
          onSubmit={(event) => {
            event.preventDefault();
            if (createValid)
              mutations.create.mutate(
                {
                  issuer: issuer.trim(),
                  clientId: clientId.trim(),
                  clientSecret: secret,
                },
                {
                  onSuccess: () => {
                    setSecret("");
                    toast.success("Identity provider connected");
                  },
                },
              );
          }}
        >
          <SecuritySettingRow
            description="The HTTPS issuer address for your OpenID Connect provider."
            htmlFor="sso-issuer"
            title="Issuer URL"
          >
            <Box className="w-full md:w-64 md:shrink-0">
              <SecurityInput
                aria-describedby="sso-issuer-description"
                disabled={busy}
                id="sso-issuer"
                maxLength={2048}
                onChange={(event) => {
                  setIssuer(event.target.value);
                }}
                placeholder="https://your-provider.example.com"
                type="url"
                value={issuer}
              />
            </Box>
          </SecuritySettingRow>
          <SecuritySettingRow
            description="The client identifier registered with your provider."
            htmlFor="sso-client-id"
            title="Client ID"
          >
            <Box className="w-full md:w-64 md:shrink-0">
              <SecurityInput
                aria-describedby="sso-client-id-description"
                disabled={busy}
                id="sso-client-id"
                maxLength={512}
                onChange={(event) => {
                  setClientId(event.target.value);
                }}
                value={clientId}
              />
            </Box>
          </SecuritySettingRow>
          <SecuritySettingRow
            description="The secret issued for this client by your provider."
            htmlFor="sso-client-secret"
            title="Client secret"
          >
            <Box className="w-full md:w-64 md:shrink-0">
              <SecurityInput
                aria-describedby="sso-client-secret-description"
                autoComplete="new-password"
                disabled={busy}
                id="sso-client-secret"
                maxLength={8192}
                onChange={(event) => {
                  setSecret(event.target.value);
                }}
                type="password"
                value={secret}
              />
            </Box>
          </SecuritySettingRow>
          <Box className="space-y-3 px-6 py-4">
            {error ? (
              <Text color="danger" role="alert">
                {error.message}
              </Text>
            ) : null}
            <Button
              disabled={!createValid || busy}
              loading={busy}
              loadingText="Connecting…"
              type="submit"
            >
              Connect identity provider
            </Button>
          </Box>
        </form>
      ) : (
        <>
          <Flex
            align="center"
            className="gap-3 px-6 py-4"
            justify="between"
            wrap
          >
            <Box className="min-w-0">
              <Text className="break-all" fontWeight="medium">
                {connection.issuer}
              </Text>
              <Text className="mt-1 break-all" color="muted">
                Client ID: {connection.clientId}
              </Text>
            </Box>
            <Badge color="tertiary" variant="outline">
              {settings.verified ? "Sign-in verified" : "Sign-in untested"}
            </Badge>
          </Flex>
          <form
            className="divide-border divide-y-[0.5px]"
            onSubmit={(event) => {
              event.preventDefault();
              mutations.update.mutate(
                {
                  enabled,
                  requireSSO: required,
                  expectedVersion: connection.version,
                  ...(rotating ? { clientSecret: secret } : {}),
                },
                {
                  onSuccess: () => {
                    setSecret("");
                    setRotating(false);
                    toast.success("SSO settings updated");
                  },
                },
              );
            }}
          >
            <SecuritySettingRow
              description="Allow linked members to sign in through this provider."
              htmlFor="sso-enabled"
              layout="inline"
              title="Enable SSO sign-in"
            >
              <Switch
                aria-describedby="sso-enabled-description"
                aria-label="Enable SSO sign-in"
                checked={enabled}
                disabled={busy}
                id="sso-enabled"
                onCheckedChange={(value) => {
                  setEnabled(value);
                  if (!value) setRequired(false);
                }}
              />
            </SecuritySettingRow>
            <SecuritySettingRow
              description="Test your sign-in first. Current sessions must then use this provider."
              htmlFor="sso-required"
              layout="inline"
              title="Require SSO for this workspace"
            >
              <Switch
                aria-describedby="sso-required-description"
                aria-label="Require SSO for this workspace"
                checked={required}
                disabled={
                  busy ||
                  !enabled ||
                  rotating ||
                  (!settings.verified && !required)
                }
                id="sso-required"
                onCheckedChange={setRequired}
              />
            </SecuritySettingRow>
            {rotating ? (
              <SecuritySettingRow
                description="SSO enforcement turns off until you test the new secret."
                htmlFor="sso-new-client-secret"
                title="New client secret"
              >
                <Box className="w-full md:w-64 md:shrink-0">
                  <SecurityInput
                    aria-describedby="sso-new-client-secret-description"
                    autoComplete="new-password"
                    disabled={busy}
                    id="sso-new-client-secret"
                    maxLength={8192}
                    onChange={(event) => {
                      setSecret(event.target.value);
                    }}
                    type="password"
                    value={secret}
                  />
                </Box>
              </SecuritySettingRow>
            ) : null}
            <Box className="space-y-3 px-6 py-4">
              {error ? (
                <Text color="danger" role="alert">
                  {error.message}
                </Text>
              ) : null}
              <Flex className="gap-3" wrap>
                <Button
                  disabled={!changed || busy || (rotating && !secret)}
                  loading={mutations.update.isPending}
                  loadingText="Saving…"
                  type="submit"
                >
                  Save SSO settings
                </Button>
                <Button
                  color="tertiary"
                  disabled={busy || !connection.enabled}
                  onClick={() => {
                    window.location.assign(settings.signInUrl);
                  }}
                  type="button"
                  variant="outline"
                >
                  Test SSO sign-in
                </Button>
                {!rotating ? (
                  <Button
                    color="tertiary"
                    disabled={busy}
                    onClick={() => {
                      setRotating(true);
                      setRequired(false);
                    }}
                    type="button"
                    variant="outline"
                  >
                    Rotate client secret
                  </Button>
                ) : (
                  <Button
                    color="tertiary"
                    disabled={busy}
                    onClick={() => {
                      setRotating(false);
                      setSecret("");
                      setRequired(connection.requireSSO);
                    }}
                    type="button"
                    variant="outline"
                  >
                    Cancel rotation
                  </Button>
                )}
              </Flex>
            </Box>
          </form>
          <SecuritySettingRow
            description="End SSO enforcement and linked sign-ins for this provider."
            title="Remove identity provider"
          >
            <Button
              color="danger"
              disabled={busy}
              onClick={() => {
                mutations.archive.reset();
                setReason("");
                setRemoving(true);
              }}
              variant="outline"
            >
              Remove identity provider
            </Button>
          </SecuritySettingRow>
        </>
      )}
      <Text className="px-6 py-4" color="muted">
        Members connect their existing account once before using SSO sign-in.
      </Text>
      <Dialog
        onOpenChange={(open) => {
          if (!busy) setRemoving(open);
        }}
        open={removing}
      >
        <Dialog.Content
          aria-busy={busy}
          className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
          hideClose={busy}
          onEscapeKeyDown={(event) => {
            if (busy) event.preventDefault();
          }}
          onInteractOutside={(event) => {
            if (busy) event.preventDefault();
          }}
        >
          <Dialog.Header className="shrink-0 px-6 py-5">
            <Dialog.Title className="text-lg">
              Remove identity provider
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-4">
            <Dialog.Description asChild className="px-0 text-base">
              <Text color="muted">
                SSO enforcement and linked sign-ins will end for this provider.
              </Text>
            </Dialog.Description>
            <SecurityInput
              disabled={busy}
              label="Reason"
              maxLength={240}
              onChange={(event) => {
                setReason(event.target.value);
              }}
              value={reason}
            />
            {mutations.archive.error ? (
              <Text color="danger" role="alert">
                {mutations.archive.error.message}
              </Text>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 gap-3 py-4" justify="end">
            <Button
              color="tertiary"
              disabled={busy}
              onClick={() => {
                setRemoving(false);
              }}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              color="danger"
              disabled={busy || !reason.trim()}
              loading={mutations.archive.isPending}
              loadingText="Removing…"
              onClick={() => {
                mutations.archive.mutate(reason.trim(), {
                  onSuccess: () => {
                    setRemoving(false);
                    toast.success("Identity provider removed");
                  },
                });
              }}
            >
              Remove provider
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};
export const SecuritySingleSignOn = () => {
  const query = useSSOSettings();
  return (
    <Box className="border-border bg-surface overflow-hidden rounded-2xl border">
      <SectionHeader
        description="Connect an OpenID Connect identity provider."
        title="Single sign-on"
      />
      {query.isPending ? (
        <Text aria-live="polite" className="px-6 py-4" color="muted">
          Loading SSO settings…
        </Text>
      ) : null}
      {query.isError ? (
        <Box className="px-6 py-4">
          <Text color="danger" role="alert">
            SSO settings could not be loaded.
          </Text>
          <Button
            className="mt-3"
            color="tertiary"
            onClick={() => void query.refetch()}
            variant="outline"
          >
            Try again
          </Button>
        </Box>
      ) : null}
      {query.data ? (
        <ConnectionForm
          key={`${query.data.connection?.id ?? "new"}:${query.data.connection?.version ?? 0}`}
          settings={query.data}
        />
      ) : null}
    </Box>
  );
};
