"use client";

import { useState, type ReactNode } from "react";
import {
  DeleteIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshIcon,
  SettingsIcon,
} from "icons";
import { toast } from "sonner";
import { Box, Button, Checkbox, Dialog, Flex, Input, Menu, Text } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { SectionHeader } from "@/modules/settings/components";
import { WEBHOOK_EVENTS, formatDeveloperDate } from "../constants";
import { useWebhookEndpoints, useWebhookMutations } from "../hooks";
import type {
  RotatedWebhookSecret,
  WebhookEndpoint,
  WebhookEventType,
} from "../types";
import { SecretDialog } from "./secret-dialog";
import { LoadError } from "./load-error";

const errorMessage = (error: unknown) =>
  error instanceof Error
    ? error.message
    : "The request could not be completed.";

const WebhookDialog = ({
  endpoint,
  isPending,
  onOpenChange,
  onSubmit,
  open,
}: {
  endpoint: WebhookEndpoint | null;
  isPending: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (input: {
    name: string;
    url: string;
    subscriptions: WebhookEventType[];
  }) => Promise<void>;
  open: boolean;
}) => {
  const [name, setName] = useState(endpoint?.name ?? "");
  const [url, setURL] = useState(endpoint?.url ?? "");
  const [subscriptions, setSubscriptions] = useState<WebhookEventType[]>(
    endpoint?.subscriptions ?? ["story.created", "story.updated"],
  );

  const reset = () => {
    setName(endpoint?.name ?? "");
    setURL(endpoint?.url ?? "");
    setSubscriptions(
      endpoint?.subscriptions ?? ["story.created", "story.updated"],
    );
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (nextOpen) reset();
    if (!isPending) onOpenChange(nextOpen);
  };

  const toggle = (eventType: WebhookEventType) => {
    setSubscriptions((current) =>
      current.includes(eventType)
        ? current.filter((value) => value !== eventType)
        : [...current, eventType],
    );
  };

  const isValid =
    name.trim().length > 0 && url.trim().length > 0 && subscriptions.length > 0;

  const submit = async () => {
    try {
      await onSubmit({
        name: name.trim(),
        url: url.trim(),
        subscriptions,
      });
      reset();
    } catch {
      // The owning section reports the API error and keeps the draft intact.
    }
  };

  return (
    <Dialog onOpenChange={handleOpenChange} open={open}>
      <Dialog.Content className="max-w-3xl" size="lg">
        <Dialog.Header className="px-6 pt-6 pb-2">
          <Dialog.Title className="text-xl">
            {endpoint ? "Edit webhook subscriptions" : "Create webhook"}
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-5 px-6 pt-2 pb-4">
          <Text color="muted">
            FortyOne signs exact request bytes and retries transient delivery
            failures. Endpoints must use public HTTPS on port 443.
          </Text>
          <Box className="space-y-5">
            <Input
              disabled={Boolean(endpoint)}
              label="Name"
              maxLength={120}
              onChange={(event) => {
                setName(event.target.value);
              }}
              placeholder="Production event processor"
              value={name}
            />
            <Input
              disabled={Boolean(endpoint)}
              label="Endpoint URL"
              maxLength={2048}
              onChange={(event) => {
                setURL(event.target.value);
              }}
              placeholder="https://example.com/webhooks/fortyone"
              type="url"
              value={url}
            />
          </Box>
          <Box>
            <Text className="mb-2" fontWeight="medium">
              Events
            </Text>
            <Box className="border-border grid grid-cols-1 overflow-hidden rounded-xl border sm:grid-cols-2">
              {WEBHOOK_EVENTS.map((eventType) => (
                <label
                  className="border-border hover:bg-state-hover flex cursor-pointer items-center gap-3 border-b p-3 last:border-b-0 sm:border-r sm:even:border-r-0 sm:[&:nth-last-child(-n+2)]:border-b-0"
                  key={eventType.value}
                >
                  <Checkbox
                    checked={subscriptions.includes(eventType.value)}
                    onCheckedChange={() => {
                      toggle(eventType.value);
                    }}
                  />
                  <Text>{eventType.label}</Text>
                </label>
              ))}
            </Box>
          </Box>
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-2">
          <Button
            color="tertiary"
            disabled={isPending}
            onClick={() => {
              handleOpenChange(false);
            }}
            variant="outline"
          >
            Cancel
          </Button>
          <Button
            disabled={!isValid}
            loading={isPending}
            onClick={() => void submit()}
          >
            {endpoint ? "Save subscriptions" : "Create webhook"}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

export const WebhookEndpoints = () => {
  const endpoints = useWebhookEndpoints();
  const mutations = useWebhookMutations();
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<WebhookEndpoint | null>(null);
  const [disableId, setDisableId] = useState<string | null>(null);
  const [secret, setSecret] = useState<{
    value: string;
    title: string;
    description: string;
  } | null>(null);
  const isPending =
    mutations.create.isPending ||
    mutations.replaceSubscriptions.isPending ||
    mutations.rotate.isPending ||
    mutations.disable.isPending;
  const endpointItems =
    endpoints.data?.pages.flatMap((page) => page.items) ?? [];

  const create = async (input: {
    name: string;
    url: string;
    subscriptions: WebhookEventType[];
  }) => {
    try {
      const value = await mutations.create.mutateAsync(input);
      setCreateOpen(false);
      setSecret({
        value: value.signingSecret,
        title: "Webhook signing secret",
        description:
          "Store this secret in your receiver's secret manager. It is required to verify every FortyOne delivery and cannot be revealed again.",
      });
      toast.success("Webhook endpoint created");
    } catch (error) {
      toast.error("Could not create webhook", {
        description: errorMessage(error),
      });
      throw error;
    }
  };

  const update = async (input: {
    name: string;
    url: string;
    subscriptions: WebhookEventType[];
  }) => {
    if (!editing) return;
    try {
      await mutations.replaceSubscriptions.mutateAsync({
        endpointId: editing.id,
        subscriptions: input.subscriptions,
      });
      setEditing(null);
      toast.success("Webhook subscriptions updated");
    } catch (error) {
      toast.error("Could not update webhook", {
        description: errorMessage(error),
      });
      throw error;
    }
  };

  const rotate = async (endpointId: string) => {
    try {
      const value: RotatedWebhookSecret =
        await mutations.rotate.mutateAsync(endpointId);
      setSecret({
        value: value.signingSecret,
        title: "New webhook signing secret",
        description: `Update your receiver before ${formatDeveloperDate(
          value.previousSecretExpiresAt,
        )}. The previous secret remains valid until then.`,
      });
      toast.success("Webhook signing secret rotated");
    } catch (error) {
      toast.error("Could not rotate webhook secret", {
        description: errorMessage(error),
      });
    }
  };

  const disable = async () => {
    if (!disableId) return;
    try {
      await mutations.disable.mutateAsync(disableId);
      setDisableId(null);
      toast.success("Webhook endpoint disabled");
    } catch (error) {
      toast.error("Could not disable webhook", {
        description: errorMessage(error),
      });
    }
  };

  let endpointContent: ReactNode = endpointItems.map((endpoint) => (
    <Flex
      align="center"
      className="gap-4 px-6 py-4"
      justify="between"
      key={endpoint.id}
      wrap
    >
      <Box className="min-w-0 flex-1">
        <Flex align="center" gap={2}>
          <Text className="font-medium">{endpoint.name}</Text>
          <Text className="capitalize" color="muted">
            {endpoint.status}
          </Text>
        </Flex>
        <Text className="mt-1 break-all" color="muted">
          {endpoint.url}
        </Text>
        <Text className="mt-1" color="muted">
          {endpoint.subscriptions.join(", ")}
          {endpoint.consecutiveFailures > 0
            ? ` · ${endpoint.consecutiveFailures} consecutive failures`
            : ""}
        </Text>
      </Box>
      {endpoint.status === "active" ? (
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Actions for ${endpoint.name}`}
              asIcon
              color="tertiary"
              disabled={isPending}
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
                  setEditing(endpoint);
                }}
              >
                <SettingsIcon className="h-[1.15rem]" />
                Manage events
              </Menu.Item>
              <Menu.Item onSelect={() => void rotate(endpoint.id)}>
                <RefreshIcon className="h-[1.15rem]" />
                Rotate secret
              </Menu.Item>
              <Menu.Item
                onSelect={() => {
                  setDisableId(endpoint.id);
                }}
              >
                <DeleteIcon className="h-[1.15rem]" />
                Disable
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      ) : null}
    </Flex>
  ));
  if (endpoints.isLoading) {
    endpointContent = (
      <Text className="px-6 py-5" color="muted">
        Loading webhooks…
      </Text>
    );
  } else if (endpoints.isError) {
    endpointContent = (
      <LoadError
        label="webhook endpoints"
        onRetry={() => void endpoints.refetch()}
      />
    );
  } else if (endpointItems.length === 0) {
    endpointContent = (
      <Text className="px-6 py-5" color="muted">
        No webhook endpoints yet.
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
        description="Deliver signed story and comment events to your systems with durable retries and replay-safe delivery IDs."
        title="Outbound webhooks"
      />
      <Box className="divide-border divide-y-[0.5px]">{endpointContent}</Box>

      {endpoints.hasNextPage ? (
        <Box className="border-border border-t-[0.5px] px-6 py-4">
          <Button
            color="tertiary"
            loading={endpoints.isFetchingNextPage}
            onClick={() => void endpoints.fetchNextPage()}
            size="sm"
            variant="outline"
          >
            Load more webhooks
          </Button>
        </Box>
      ) : null}

      <WebhookDialog
        endpoint={null}
        isPending={mutations.create.isPending}
        onOpenChange={setCreateOpen}
        onSubmit={create}
        open={createOpen}
      />
      <WebhookDialog
        endpoint={editing}
        isPending={mutations.replaceSubscriptions.isPending}
        key={editing?.id ?? "no-endpoint"}
        onOpenChange={(open) => {
          if (!open) setEditing(null);
        }}
        onSubmit={update}
        open={Boolean(editing)}
      />
      <SecretDialog
        description={secret?.description ?? ""}
        label="Signing secret"
        onClose={() => {
          setSecret(null);
        }}
        open={Boolean(secret)}
        secret={secret?.value ?? ""}
        title={secret?.title ?? "Webhook signing secret"}
      />
      <ConfirmDialog
        confirmPhrase="disable"
        confirmText="Disable webhook"
        description="New events will no longer be queued for this endpoint. Existing delivery state remains available for audit and recovery."
        isLoading={mutations.disable.isPending}
        isOpen={Boolean(disableId)}
        onClose={() => {
          setDisableId(null);
        }}
        onConfirm={() => void disable()}
        title="Disable webhook endpoint?"
      />
    </Box>
  );
};
