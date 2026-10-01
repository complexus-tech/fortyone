"use client";

import { useState } from "react";
import { Badge, Box, Button, Dialog, Flex, Input, Switch, Text } from "ui";
import { SectionHeader } from "@/modules/settings/components/section-header";
import type { SSOSettings } from "./sso-api";
import { useSSOMutations, useSSOSettings } from "./sso-api";

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
    <Box className="space-y-5 px-6 pt-5 pb-6">
      <Box className="bg-surface-elevated border-border rounded-xl border p-4">
        <Text fontWeight="medium">Identity provider callback URL</Text>
        <Text className="mt-2 break-all select-all">
          {settings.callbackUrl}
        </Text>
      </Box>
      {!connection ? (
        <form
          className="max-w-2xl space-y-4"
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
                  },
                },
              );
          }}
        >
          <Input
            disabled={busy}
            label="Issuer URL"
            maxLength={2048}
            onChange={(event) => {
              setIssuer(event.target.value);
            }}
            placeholder="https://your-provider.example.com"
            type="url"
            value={issuer}
          />
          <Input
            disabled={busy}
            label="Client ID"
            maxLength={512}
            onChange={(event) => {
              setClientId(event.target.value);
            }}
            value={clientId}
          />
          <Input
            autoComplete="new-password"
            disabled={busy}
            label="Client secret"
            maxLength={8192}
            onChange={(event) => {
              setSecret(event.target.value);
            }}
            type="password"
            value={secret}
          />
          {error ? (
            <Text color="danger" role="alert">
              {error.message}
            </Text>
          ) : null}
          <Button disabled={!createValid || busy} loading={busy} type="submit">
            {busy ? "Connecting…" : "Connect identity provider"}
          </Button>
        </form>
      ) : (
        <>
          <Flex align="center" className="gap-3" justify="between" wrap>
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
            className="max-w-2xl space-y-5"
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
                  },
                },
              );
            }}
          >
            <Flex align="center" className="gap-5" justify="between">
              <Box>
                <label className="font-medium" htmlFor="sso-enabled">
                  Enable SSO sign-in
                </label>
                <Text className="mt-1" color="muted">
                  Allow linked members to sign in through this provider.
                </Text>
              </Box>
              <Switch
                aria-label="Enable SSO sign-in"
                checked={enabled}
                disabled={busy}
                id="sso-enabled"
                onCheckedChange={(value) => {
                  setEnabled(value);
                  if (!value) setRequired(false);
                }}
              />
            </Flex>
            <Flex align="center" className="gap-5" justify="between">
              <Box>
                <label className="font-medium" htmlFor="sso-required">
                  Require SSO for this workspace
                </label>
                <Text className="mt-1" color="muted">
                  Test your sign-in first. Current sessions must then use this
                  provider.
                </Text>
              </Box>
              <Switch
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
            </Flex>
            {rotating ? (
              <Input
                autoComplete="new-password"
                disabled={busy}
                label="New client secret"
                maxLength={8192}
                onChange={(event) => {
                  setSecret(event.target.value);
                }}
                type="password"
                value={secret}
              />
            ) : null}
            {rotating ? (
              <Text color="muted">
                SSO enforcement turns off until you test the new secret.
              </Text>
            ) : null}
            {error ? (
              <Text color="danger" role="alert">
                {error.message}
              </Text>
            ) : null}
            <Flex className="gap-3" wrap>
              <Button
                disabled={!changed || busy || (rotating && !secret)}
                loading={mutations.update.isPending}
                type="submit"
              >
                {mutations.update.isPending ? "Saving…" : "Save SSO settings"}
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
          </form>
          <Box className="border-border border-t pt-5">
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
          </Box>
        </>
      )}
      <Text color="muted">
        Members connect their existing account once before using SSO sign-in.
      </Text>
      <Dialog
        onOpenChange={(open) => {
          if (!busy) setRemoving(open);
        }}
        open={removing}
      >
        <Dialog.Content
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
            <Dialog.Description asChild className="px-0 text-base leading-6">
              <Text color="muted">
                SSO enforcement and linked sign-ins will end for this provider.
              </Text>
            </Dialog.Description>
            <Input
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
              onClick={() => {
                mutations.archive.mutate(reason.trim(), {
                  onSuccess: () => {
                    setRemoving(false);
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
        <Text aria-live="polite" className="px-6 pt-5 pb-6" color="muted">
          Loading SSO settings…
        </Text>
      ) : null}
      {query.isError ? (
        <Box className="px-6 pt-5 pb-6">
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
