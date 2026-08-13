"use client";

import { useMemo, useState } from "react";
import { CheckIcon, CopyIcon, DeleteIcon, PlusIcon } from "icons";
import { Box, Button, Flex, Input, Switch, Text } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import type {
  FeedbackWidgetSettings,
  FeedbackWidgetSigningSecret,
} from "./types";
import {
  useCreateFeedbackWidgetSecretMutation,
  useFeedbackWidgetSettings,
  useRotateFeedbackWidgetSecretMutation,
  useUpdateFeedbackWidgetSettingsMutation,
} from "./hooks";
import { normalizeFeedbackWidgetOrigin } from "./widget-origin";

const CopyButton = ({ label, value }: { label: string; value: string }) => {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      window.setTimeout(() => {
        setCopied(false);
      }, 1800);
    } catch {
      setCopied(false);
    }
  };

  return (
    <Button
      aria-label={label}
      asIcon
      color="tertiary"
      onClick={() => {
        void copy();
      }}
      size="sm"
      variant="naked"
    >
      {copied ? <CheckIcon className="h-4" /> : <CopyIcon className="h-4" />}
    </Button>
  );
};

const SigningSecretReveal = ({
  secret,
}: {
  secret: FeedbackWidgetSigningSecret;
}) => (
  <Box className="border-warning/30 bg-warning/5 mt-4 rounded-xl border p-4">
    <Text className="text-sm" fontWeight="semibold">
      Save this signing secret now
    </Text>
    <Text className="mt-1 max-w-2xl text-sm leading-5" color="muted">
      It is shown only once. Store it in your application&apos;s server-side
      environment and never place it in browser code or the widget snippet.
    </Text>
    <Flex
      align="center"
      className="border-border bg-background mt-3 min-w-0 rounded-lg border px-3 py-1.5"
      gap={2}
    >
      <code className="min-w-0 flex-1 overflow-x-auto text-xs whitespace-nowrap">
        {secret.signingSecret}
      </code>
      <CopyButton
        label="Copy widget signing secret"
        value={secret.signingSecret}
      />
    </Flex>
  </Box>
);

const WidgetSecurityEditor = ({
  initialSettings,
  portalId,
}: {
  initialSettings: FeedbackWidgetSettings;
  portalId: string;
}) => {
  const [enabled, setEnabled] = useState(initialSettings.enabled);
  const [allowedOrigins, setAllowedOrigins] = useState(
    initialSettings.allowedOrigins,
  );
  const [originInput, setOriginInput] = useState("");
  const [originError, setOriginError] = useState("");
  const [accessError, setAccessError] = useState("");
  const [revealedSecret, setRevealedSecret] =
    useState<FeedbackWidgetSigningSecret | null>(null);
  const [confirmRotation, setConfirmRotation] = useState(false);
  const updateMutation = useUpdateFeedbackWidgetSettingsMutation(portalId);
  const createSecretMutation = useCreateFeedbackWidgetSecretMutation(
    portalId,
    setRevealedSecret,
  );
  const rotateSecretMutation = useRotateFeedbackWidgetSecretMutation(
    portalId,
    (result) => {
      setRevealedSecret(result);
      setConfirmRotation(false);
    },
  );
  const isDirty = useMemo(
    () =>
      enabled !== initialSettings.enabled ||
      allowedOrigins.join("\n") !== initialSettings.allowedOrigins.join("\n"),
    [allowedOrigins, enabled, initialSettings],
  );
  const settings = revealedSecret ?? initialSettings;

  const addOrigin = () => {
    try {
      const origin = normalizeFeedbackWidgetOrigin(originInput);
      if (allowedOrigins.includes(origin)) {
        setOriginError("This origin is already allowed");
        return;
      }
      setAllowedOrigins((current) => [...current, origin]);
      setOriginInput("");
      setOriginError("");
      setAccessError("");
    } catch (error) {
      setOriginError(
        error instanceof Error ? error.message : "Enter a valid origin",
      );
    }
  };

  const save = async () => {
    if (enabled && allowedOrigins.length === 0) {
      setOriginError("Add at least one origin before enabling the widget");
      return;
    }
    if (enabled && !settings.hasSigningSecret) {
      setAccessError(
        "Create a signing secret before enabling signed widget access.",
      );
      return;
    }
    setAccessError("");
    await updateMutation.mutateAsync({ allowedOrigins, enabled });
  };

  return (
    <Box className="space-y-6">
      <Flex align="center" gap={4} justify="between">
        <Box>
          <Text fontWeight="medium">Enable embeds</Text>
          <Text className="mt-1 max-w-xl text-sm" color="muted">
            Only pages on the exact origins below can load this feedback widget
            or exchange a signed customer identity.
          </Text>
        </Box>
        <Switch
          aria-label="Enable feedback widget"
          checked={enabled}
          onCheckedChange={(checked) => {
            setEnabled(checked);
            setAccessError("");
          }}
        />
      </Flex>

      <Box className="border-border/70 border-t pt-5">
        <Text fontWeight="medium">Allowed origins</Text>
        <Text className="mt-1 text-sm" color="muted">
          Include the scheme and optional port, for example
          https://app.example.com. Paths and wildcards are not accepted.
        </Text>
        <Flex align="start" className="mt-3 flex-col gap-2 sm:flex-row">
          <Box className="w-full max-w-xl">
            <Input
              aria-invalid={Boolean(originError)}
              onChange={(event) => {
                setOriginInput(event.target.value);
                setOriginError("");
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  addOrigin();
                }
              }}
              placeholder="https://app.example.com"
              value={originInput}
            />
            {originError ? (
              <Text className="mt-1.5 text-sm text-red-600 dark:text-red-400">
                {originError}
              </Text>
            ) : null}
          </Box>
          <Button
            color="tertiary"
            leftIcon={<PlusIcon className="h-4" />}
            onClick={addOrigin}
          >
            Add origin
          </Button>
        </Flex>
        {allowedOrigins.length > 0 ? (
          <Box className="mt-3 space-y-2">
            {allowedOrigins.map((origin) => (
              <Flex
                align="center"
                className="border-border/70 bg-background max-w-2xl rounded-xl border px-3 py-2"
                gap={2}
                justify="between"
                key={origin}
              >
                <code className="min-w-0 truncate text-xs">{origin}</code>
                <Button
                  aria-label={`Remove ${origin}`}
                  asIcon
                  color="tertiary"
                  onClick={() => {
                    setAllowedOrigins((current) =>
                      current.filter((candidate) => candidate !== origin),
                    );
                  }}
                  size="sm"
                  variant="naked"
                >
                  <DeleteIcon className="h-4" />
                </Button>
              </Flex>
            ))}
          </Box>
        ) : null}
        <Button
          className="mt-4"
          color="primary"
          disabled={!isDirty || updateMutation.isPending}
          loading={updateMutation.isPending}
          onClick={() => {
            void save();
          }}
          size="sm"
        >
          Save widget access
        </Button>
        {accessError ? (
          <Text className="mt-2 text-sm text-red-600 dark:text-red-400">
            {accessError}
          </Text>
        ) : null}
      </Box>

      <Box className="border-border/70 border-t pt-5">
        <Flex
          align="start"
          className="flex-col gap-4 sm:flex-row"
          justify="between"
        >
          <Box>
            <Text fontWeight="medium">Signed customer identity</Text>
            <Text className="mt-1 max-w-xl text-sm leading-5" color="muted">
              Your backend signs a short-lived identity assertion. The widget
              exchanges it for a portal-scoped contributor session, so the
              signing secret never reaches browser storage.
            </Text>
          </Box>
          {settings.hasSigningSecret ? (
            <Button
              color="tertiary"
              onClick={() => {
                setConfirmRotation(true);
              }}
              size="sm"
            >
              Rotate secret
            </Button>
          ) : (
            <Button
              color="tertiary"
              loading={createSecretMutation.isPending}
              onClick={() => {
                createSecretMutation.mutate();
              }}
              size="sm"
            >
              Create signing secret
            </Button>
          )}
        </Flex>
        <Box className="mt-4 grid gap-3 sm:grid-cols-2">
          <Box className="border-border/70 bg-background rounded-xl border p-3">
            <Text className="text-xs" color="muted">
              Widget key ID
            </Text>
            <Flex align="center" className="mt-1 min-w-0" gap={2}>
              <code className="min-w-0 flex-1 truncate text-xs">
                {settings.widgetKeyId}
              </code>
              <CopyButton
                label="Copy widget key ID"
                value={settings.widgetKeyId}
              />
            </Flex>
          </Box>
          <Box className="border-border/70 bg-background rounded-xl border p-3">
            <Text className="text-xs" color="muted">
              Active secret version
            </Text>
            <Text className="mt-1 text-sm" fontWeight="medium">
              {settings.hasSigningSecret
                ? `Version ${settings.signingSecretVersion}`
                : "Not configured"}
            </Text>
          </Box>
        </Box>
        {settings.previousVersionExpiresAt ? (
          <Text className="mt-3 text-sm" color="muted">
            The previous secret remains valid until{" "}
            {new Date(settings.previousVersionExpiresAt).toLocaleString()} so
            you can deploy the replacement safely.
          </Text>
        ) : null}
        {revealedSecret ? (
          <SigningSecretReveal secret={revealedSecret} />
        ) : null}
      </Box>

      <ConfirmDialog
        confirmPhrase="rotate"
        confirmText="Rotate secret"
        description="A new secret will be shown once. The current secret remains valid only during the grace period, so update your backend before that period ends."
        isLoading={rotateSecretMutation.isPending}
        isOpen={confirmRotation}
        loadingText="Rotating…"
        onClose={() => {
          setConfirmRotation(false);
        }}
        onConfirm={() => {
          rotateSecretMutation.mutate();
        }}
        title="Rotate widget signing secret?"
      />
    </Box>
  );
};

export const WidgetSecuritySettings = ({ portalId }: { portalId: string }) => {
  const query = useFeedbackWidgetSettings(portalId);

  if (query.isLoading) {
    return (
      <Box className="border-border/70 bg-surface-muted/30 rounded-xl border p-4">
        <Text color="muted">Loading widget access settings…</Text>
      </Box>
    );
  }
  if (query.error || !query.data) {
    return (
      <Box className="border-border/70 bg-surface-muted/30 rounded-xl border p-4">
        <Text className="text-sm text-red-600 dark:text-red-400">
          {query.error?.message ?? "Widget access settings are unavailable"}
        </Text>
        <Button
          className="mt-3"
          color="tertiary"
          onClick={() => {
            void query.refetch();
          }}
          size="sm"
        >
          Try again
        </Button>
      </Box>
    );
  }

  return (
    <WidgetSecurityEditor
      initialSettings={query.data}
      key={`${query.data.signingSecretVersion}:${query.data.allowedOrigins.join(",")}:${query.data.enabled}`}
      portalId={portalId}
    />
  );
};
